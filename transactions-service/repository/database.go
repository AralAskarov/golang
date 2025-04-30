package repository

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

func NewDatabase(connectionString string) (*sql.DB, error) {
    db, err := sql.Open("mysql", connectionString)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}
