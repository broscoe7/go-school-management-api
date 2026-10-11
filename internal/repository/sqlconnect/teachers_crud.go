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

func GetTeachersDBHandler(teachers []models.Teacher, r *http.Request, page, limit int) ([]models.Teacher, int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, 0, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := `
		SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1
	`
	var args []any
	query, args = utils.AddFilters(r, query, args)
	offset := (page - 1) * limit
	query += "LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, utils.ErrorHandler(err, "Database query error")
	}
	defer rows.Close()
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			// http.Error(w, "Error scanning database results", http.StatusInternalServerError)
			return nil, 0, err
		}
		teachers = append(teachers, teacher)
	}

	var totalTeachers int
	err = db.QueryRow(`SELECT COUNT(*) FROM teachers`).Scan(&totalTeachers)
	if err != nil {
		utils.ErrorHandler(err, "")
		totalTeachers = 0
	}
	return teachers, totalTeachers, nil
}

func GetTeacherByID(id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var teacher models.Teacher
	query := `SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?`
	err = db.QueryRow(query, id).Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error retrieving data")
	}
	return teacher, nil
}

func AddTeachersDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	stmt, err := db.Prepare(utils.GenerateInsertQuery("teachers", models.Teacher{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error preparing query")
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, teacher := range newTeachers {
		values := utils.GetStructValues(teacher)
		res, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into database")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into database")
		}
		teacher.ID = int(lastID)
		addedTeachers[i] = teacher
	}
	return addedTeachers, nil
}

func UpdateTeacherDBHandler(id int, updatedTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID,
		&existingTeacher.FirstName,
		&existingTeacher.LastName,
		&existingTeacher.Email,
		&existingTeacher.Class,
		&existingTeacher.Subject,
	)
	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, "Record not found")
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error retrieving data")
	}
	updatedTeacher.ID = existingTeacher.ID
	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		updatedTeacher.FirstName,
		updatedTeacher.LastName,
		updatedTeacher.Email,
		updatedTeacher.Class,
		updatedTeacher.Subject,
		updatedTeacher.ID,
	)
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error updating record")
	}
	return updatedTeacher, nil
}

func PatchTeachers(updates []map[string]any) error {
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

		var teacherFromDB models.Teacher
		err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(
			&teacherFromDB.ID,
			&teacherFromDB.FirstName,
			&teacherFromDB.LastName,
			&teacherFromDB.Email,
			&teacherFromDB.Class,
			&teacherFromDB.Subject,
		)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return utils.ErrorHandler(err, "Record not found")
		} else if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error retrieving data")
		}

		teacherVal := reflect.ValueOf(&teacherFromDB).Elem()
		teacherType := teacherVal.Type()
		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if strings.Split(field.Tag.Get("json"), ",")[0] == k {
					fieldVal := teacherVal.Field(i)
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
		_, err = tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
			teacherFromDB.FirstName,
			teacherFromDB.LastName,
			teacherFromDB.Email,
			teacherFromDB.Class,
			teacherFromDB.Subject,
			teacherFromDB.ID,
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

func PatchTeacherDB(id int, updates map[string]any) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID,
		&existingTeacher.FirstName,
		&existingTeacher.LastName,
		&existingTeacher.Email,
		&existingTeacher.Class,
		&existingTeacher.Subject,
	)
	if err == sql.ErrNoRows {
		return models.Teacher{}, utils.ErrorHandler(err, "Record not found")
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error retrieving data")
	}

	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()
	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)
			if strings.Split(field.Tag.Get("json"), ",")[0] == k {
				if teacherVal.Field(i).CanSet() {
					teacherVal.Field(i).Set(reflect.ValueOf(v).Convert(teacherVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		existingTeacher.FirstName,
		existingTeacher.LastName,
		existingTeacher.Email,
		existingTeacher.Class,
		existingTeacher.Subject,
		existingTeacher.ID,
	)
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error updating record")
	}
	return existingTeacher, nil
}

func DeleteTeacherDB(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
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

func DeleteTeachersDB(ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error starting transaction")
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
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

func GetStudentsByTeacherIDFromDB(teacherID int, students []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := `
		SELECT id, first_name, last_name, email, class
		FROM students WHERE class = (SELECT class FROM teachers WHERE id = ?)
	`
	rows, err := db.Query(query, teacherID)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error retrieving records")
	}
	defer rows.Close()
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error retrieving records")
		}
		students = append(students, student)
		err = rows.Err()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error retrieving records")
		}
	}
	return students, nil
}

func GetStudentsCountByTeacherIDFromDB(teacherID, studentCount int) (int, error) {
	db, err := ConnectDB()
	if err != nil {
		return 0, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := `
		SELECT COUNT(*) FROM students
		WHERE class = (SELECT class FROM teachers WHERE id = ?)
	`
	err = db.QueryRow(query, teacherID).Scan(&studentCount)
	if err != nil {
		return 0, utils.ErrorHandler(err, "Error retrieving data")
	}
	return studentCount, nil
}
