package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // postgres driver
)

var db *sql.DB

func init() {
	//err := ConnetctDB(os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	err := ConnetctDB("forum_db_user", "qwerty", "localhost", "forum_db")
	if err != nil {
		log.Fatalf("cant open database connection: %s", err.Description)
	}
}

// ConnetctDB создаёт новый коннект к базе и заворачивает в контроллер
func ConnetctDB(dbUser, dbPass, dbHost, dbName string) *Error {
	newDB, err := sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s/%s", dbUser, dbPass, dbHost, dbName))
	if err != nil {
		return NewError(InternalDatabase, err.Error(), "")
	}

	if err := newDB.Ping(); err != nil {
		return NewError(InternalDatabase, err.Error(), "")
	}

	db = newDB

	return nil
}

// GetDB возвращает указатель на контроллер,
// для дополнительных запросов, не включенных в модели;
func GetDB() *sql.DB {
	return db
}
