package db

import (
	"database/sql"
	"school-api/internal/repositeries/sqlconnect"
)

type DB struct {
	*sql.DB
}

func New() (*DB, error) {
	conn, err := sqlconnect.ConnectDB()
	if err != nil {
		return nil, err
	}
	return &DB{DB: conn}, nil
}

func (d *DB) Close() {
	d.DB.Close()
}
