package models

import (
	"database/sql"
	"fmt"

	// пустой импорт для работы с бд

	_ "github.com/lib/pq"
)

type queryer interface {
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
}

var db *sql.DB

//err := ConnetctDB(os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

// ConnetctDB создаёт новый коннект к базе и заворачивает в контроллер
func ConnetctDB(dbUser, dbPass, dbHost, dbName string) *Error {
	newDB, err := sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s/%s", dbUser, dbPass, dbHost, dbName))
	if err != nil {
		return NewError(InternalDatabase, err.Error())
	}

	if err := newDB.Ping(); err != nil {
		return NewError(InternalDatabase, err.Error())
	}

	db = newDB
	return nil
}

// GetDB возвращает указатель на контроллер,
// для дополнительных запросов, не включенных в модели;
func GetDB() *sql.DB {
	return db
}
