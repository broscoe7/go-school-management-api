package sqlconnect

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/broscoe7/go-school-api/internal/models"
	"github.com/broscoe7/go-school-api/pkg/utils"
	"gopkg.in/gomail.v2"
)

func GetExecsDBHandler(execs []models.Exec, r *http.Request) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := `
		SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE 1=1
	`
	var args []any
	query, args = utils.AddFilters(r, query, args)
	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database query error")
	}
	defer rows.Close()
	for rows.Next() {
		var exec models.Exec
		err := rows.Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.Username, &exec.UserCreatedAt, &exec.InactiveStatus, &exec.Role)
		if err != nil {
			// http.Error(w, "Error scanning database results", http.StatusInternalServerError)
			return nil, err
		}
		execs = append(execs, exec)
	}
	return execs, nil
}

func GetExecByID(id int) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var exec models.Exec
	query := `SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE id = ?`
	err = db.QueryRow(query, id).Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.Username, &exec.UserCreatedAt, &exec.InactiveStatus, &exec.Role)
	if err == sql.ErrNoRows {
		return models.Exec{}, utils.ErrorHandler(err, "Exec not found")
	} else if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error retrieving data")
	}
	return exec, nil
}

func GetExecByUsername(username string) (*models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()
	user := &models.Exec{}
	err = db.QueryRow(`SELECT id, first_name, last_name, email, username, password, inactive_status, role
		FROM execs WHERE username = ?`, username).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.InactiveStatus,
		&user.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrorHandler(err, "User not found")
		}
		return nil, utils.ErrorHandler(err, "Database query error")
	}
	return user, nil
}

func AddExecsDBHandler(newExecs []models.Exec) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	stmt, err := db.Prepare(utils.GenerateInsertQuery("execs", models.Exec{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error preparing query")
	}
	defer stmt.Close()

	addedExecs := make([]models.Exec, len(newExecs))
	for i, exec := range newExecs {
		exec.Password, err = utils.HashPassword(exec.Password)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error inserting data into database")
		}

		values := utils.GetStructValues(exec)
		res, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into database")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into database")
		}
		exec.ID = int(lastID)
		addedExecs[i] = exec
	}
	return addedExecs, nil
}

func PatchExecs(updates []map[string]any) error {
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

		var execFromDB models.Exec
		err = db.QueryRow("SELECT id, first_name, last_name, email, username FROM execs WHERE id = ?", id).Scan(
			&execFromDB.ID,
			&execFromDB.FirstName,
			&execFromDB.LastName,
			&execFromDB.Email,
			&execFromDB.Username,
		)
		if err == sql.ErrNoRows {
			tx.Rollback()
			return utils.ErrorHandler(err, "Record not found")
		} else if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error retrieving data")
		}

		execVal := reflect.ValueOf(&execFromDB).Elem()
		execType := execVal.Type()
		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < execVal.NumField(); i++ {
				field := execType.Field(i)
				if strings.Split(field.Tag.Get("json"), ",")[0] == k {
					fieldVal := execVal.Field(i)
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
		_, err = tx.Exec("UPDATE execs SET first_name = ?, last_name = ?, email = ?, username = ? WHERE id = ?",
			execFromDB.FirstName,
			execFromDB.LastName,
			execFromDB.Email,
			execFromDB.Username,
			execFromDB.ID,
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

func PatchExecDB(id int, updates map[string]any) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var existingExec models.Exec
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, username FROM execs WHERE id = ?", id).Scan(
		&existingExec.ID,
		&existingExec.FirstName,
		&existingExec.LastName,
		&existingExec.Email,
		&existingExec.Username,
	)
	if err == sql.ErrNoRows {
		return models.Exec{}, utils.ErrorHandler(err, "Record not found")
	} else if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error retrieving data")
	}

	execVal := reflect.ValueOf(&existingExec).Elem()
	execType := execVal.Type()
	for k, v := range updates {
		for i := 0; i < execVal.NumField(); i++ {
			field := execType.Field(i)
			if strings.Split(field.Tag.Get("json"), ",")[0] == k {
				if execVal.Field(i).CanSet() {
					execVal.Field(i).Set(reflect.ValueOf(v).Convert(execVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE execs SET first_name = ?, last_name = ?, email = ?, username = ? WHERE id = ?",
		existingExec.FirstName,
		existingExec.LastName,
		existingExec.Email,
		existingExec.Username,
		existingExec.ID,
	)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error updating record")
	}
	return existingExec, nil
}

func DeleteExecDB(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM execs WHERE id = ?", id)
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

func UpdatePasswordInDB(userID int, currentPassword, newPassword string) (string, error) {
	db, err := ConnectDB()
	if err != nil {
		return "", utils.ErrorHandler(err, "database connection error")
	}
	defer db.Close()

	var username string
	var userPassword string
	var userRole string
	err = db.QueryRow("SELECT username, password, role FROM execs WHERE id = ?", userID).Scan(&username, &userPassword, &userRole)
	if err != nil {
		return "", utils.ErrorHandler(err, "User not found")
	}

	err = utils.VerifyPassword(currentPassword, userPassword)
	if err != nil {
		return "", utils.ErrorHandler(err, "The password does not match the current password on file")
	}
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return "", utils.ErrorHandler(err, "internal error")
	}
	currentTime := time.Now().Format(time.RFC3339)
	_, err = db.Exec("UPDATE execs SET password = ?, password_changed_at = ? WHERE id = ?", hashedPassword, currentTime, userID)
	if err != nil {
		return "", utils.ErrorHandler(err, "failed to update password")
	}
	// after successful update we send the user a new JWT token
	token, err := utils.SignToken(userID, username, userRole)
	if err != nil {
		return "", utils.ErrorHandler(err, "password updated but could not create new token")
	}
	return token, nil
}

func ForgotPasswordDBHandler(email string) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	// find user by email
	var exec models.Exec
	err = db.QueryRow(`SELECT id FROM execs WHERE email = ?`, email).Scan(&exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "User not found")
	}

	// create password reset token and store in db
	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	}
	minutes := time.Duration(duration)
	expiry := time.Now().Add(minutes * time.Minute).Format(time.RFC3339)
	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	}
	token := hex.EncodeToString(tokenBytes)
	hashedToken := sha256.Sum256(tokenBytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	_, err = db.Exec(`UPDATE execs SET password_reset_token = ?, password_token_expires = ? WHERE id = ?`, hashedTokenString, expiry, exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	}

	// send reset email
	resetURL := fmt.Sprintf("https://localhost:3000/execs/password/reset/%s", token)
	message := fmt.Sprintf("Forgot your password? Reset it using the following link: \n%s\nIf you didn't request a password reset, please ignore this email. This link is only valid for %d minutes.", resetURL, int(minutes))
	m := gomail.NewMessage()
	m.SetHeader("From", "schooladmin@school.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your password reset link")
	m.SetBody("text/plain", message)
	d := gomail.NewDialer("localhost", 1025, "", "")
	err = d.DialAndSend(m)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	}
	return nil
}

func ResetPasswordDBHandler(token, newPassword string) error {
	// hash token
	bytes, err := hex.DecodeString(token)
	if err != nil {
		return utils.ErrorHandler(err, "internal server error")
	}
	hashedToken := sha256.Sum256(bytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "database connection error")
	}
	defer db.Close()

	// check token against database records
	var user models.Exec
	query := `SELECT id, email FROM execs WHERE password_reset_token = ? AND password_token_expires > ?`
	err = db.QueryRow(query, hashedTokenString, time.Now().Format(time.RFC3339)).Scan(&user.ID, &user.Email)
	if err != nil {
		return utils.ErrorHandler(err, "Invalid or expired reset token")
	}

	// hash the new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return utils.ErrorHandler(err, "internal server error")
	}

	// save new password in db
	updateQuery := `UPDATE execs
		SET password = ?, password_reset_token = NULL, password_token_expires = NULL, password_changed_at = ?
		WHERE id = ?`
	_, err = db.Exec(updateQuery, hashedPassword, time.Now().Format(time.RFC3339), user.ID)
	if err != nil {
		return utils.ErrorHandler(err, "internal server error")
	}
	return nil
}
