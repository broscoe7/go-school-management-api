package sqlconnect

import (
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/broscoe7/go-school-api/internal/models"
	"github.com/broscoe7/go-school-api/pkg/utils"
)

func GetStudentsDBHandler(students []models.Student, r *http.Request, page, limit int) ([]models.Student, int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, 0, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := `
		SELECT id, first_name, last_name, email, class FROM students WHERE 1=1
	`
	var args []any
	// add filters
	query, args = utils.AddFilters(r, query, args)

	// add pagination
	offset := (page - 1) * limit
	query += "LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// add sorting
	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, utils.ErrorHandler(err, "Database query error")
	}
	defer rows.Close()
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, 0, err
		}
		students = append(students, student)
	}

	// get total count but do not throw error if something goes wrong
	var totalStudents int
	err = db.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&totalStudents)
	if err != nil {
		utils.ErrorHandler(err, "")
		totalStudents = 0
	}

	return students, totalStudents, nil
}

func GetStudentByID(id int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var student models.Student
	query := `SELECT id, first_name, last_name, email, class FROM students WHERE id = ?`
	err = db.QueryRow(query, id).Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "Student not found")
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error retrieving data")
	}
	return student, nil
}

func AddStudentsDBHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error preparing query")
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, student := range newStudents {
		values := utils.GetStructValues(student)
		res, err := stmt.Exec(values...)
		if err != nil {
			if strings.Contains(err.Error(), "a foreign key constraint fails") {
				return nil, utils.ErrorHandler(err, "Class or class teacher does not exist")
			}
			return nil, utils.ErrorHandler(err, "Error inserting data into database")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into database")
		}
		student.ID = int(lastID)
		addedStudents[i] = student
	}
	return addedStudents, nil
}

func UpdateStudentDBHandler(id int, updatedStudent models.Student) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(
		&existingStudent.ID,
		&existingStudent.FirstName,
		&existingStudent.LastName,
		&existingStudent.Email,
		&existingStudent.Class,
	)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "Record not found")
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error retrieving data")
	}
	updatedStudent.ID = existingStudent.ID
	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
		updatedStudent.FirstName,
		updatedStudent.LastName,
		updatedStudent.Email,
		updatedStudent.Class,
		updatedStudent.ID,
	)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error updating record")
	}
	return updatedStudent, nil
}

func PatchStudents(updates []map[string]any) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "Error starting transaction")
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return utils.ErrorHandler(err, "Invalid id")
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Invalid id")
		}

		var studentFromDB models.Student
		err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(
			&studentFromDB.ID,
			&studentFromDB.FirstName,
			&studentFromDB.LastName,
			&studentFromDB.Email,
			&studentFromDB.Class,
		)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return utils.ErrorHandler(err, "Record not found")
		} else if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error retrieving data")
		}

		studentVal := reflect.ValueOf(&studentFromDB).Elem()
		studentType := studentVal.Type()
		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < studentVal.NumField(); i++ {
				field := studentType.Field(i)
				if strings.Split(field.Tag.Get("json"), ",")[0] == k {
					fieldVal := studentVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							return fmt.Errorf("cannot convert %v to %v", val.Type(), fieldVal.Type())
						}
					}
					break
				}
			}
		}
		_, err = tx.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
			studentFromDB.FirstName,
			studentFromDB.LastName,
			studentFromDB.Email,
			studentFromDB.Class,
			studentFromDB.ID,
		)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error updating record")
		}
	}
	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "Error updating record")
	}
	return nil
}

func PatchStudentDB(id int, updates map[string]any) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(
		&existingStudent.ID,
		&existingStudent.FirstName,
		&existingStudent.LastName,
		&existingStudent.Email,
		&existingStudent.Class,
	)
	if err == sql.ErrNoRows {
		return models.Student{}, utils.ErrorHandler(err, "Record not found")
	} else if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error retrieving data")
	}

	studentVal := reflect.ValueOf(&existingStudent).Elem()
	studentType := studentVal.Type()
	for k, v := range updates {
		for i := 0; i < studentVal.NumField(); i++ {
			field := studentType.Field(i)
			if strings.Split(field.Tag.Get("json"), ",")[0] == k {
				if studentVal.Field(i).CanSet() {
					studentVal.Field(i).Set(reflect.ValueOf(v).Convert(studentVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
		existingStudent.FirstName,
		existingStudent.LastName,
		existingStudent.Email,
		existingStudent.Class,
		existingStudent.ID,
	)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error updating record")
	}
	return existingStudent, nil
}

func DeleteStudentDB(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting record")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting record")
	}
	if rowsAffected == 0 {
		return utils.ErrorHandler(err, "Record not found")
	}
	return nil
}

func DeleteStudentsDB(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error starting transaction")
	}

	stmt, err := tx.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "Error preparing delete statement")
	}
	defer stmt.Close()

	var deletedIDs = []int{}
	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error deleting records")
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error retrieving delete result")
		}
		if rowsAffected == 0 {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, fmt.Sprintf("ID %d not found", id))
		}
		deletedIDs = append(deletedIDs, id)
	}
	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error deleting records")
	}

	if len(deletedIDs) < 1 {
		return nil, utils.ErrorHandler(err, "Records not found")
	}
	return deletedIDs, nil
}
