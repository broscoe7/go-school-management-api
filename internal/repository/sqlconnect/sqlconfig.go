package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() (*sql.DB, error) {
	db_user := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("HOST")
	db_port := os.Getenv("DB_PORT")

	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", db_user, db_password, host, db_port, dbname)
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to MySQL database")
	return db, nil
}
