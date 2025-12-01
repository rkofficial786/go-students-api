package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	Conn *sql.DB
}

func ConnectDB() (*sql.DB, error) {

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(localhost:3306)/%s", user,
		password,
		dbName,
	)
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		return nil, err
	}
	fmt.Println("Connect to database successfully -", dbName)

	return db, nil
}


