package models

import (
	"github.com/jackc/pgx"
)

type queryer interface {
	QueryRow(string, ...interface{}) *pgx.Row
	Query(string, ...interface{}) (*pgx.Rows, error)
}

type executer interface {
	Exec(string, ...interface{}) (pgx.CommandTag, error)
}

type exequeryer interface {
	queryer
	executer
}

var db *pgx.ConnPool

//err := ConnetctDB(os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

// ConnetctDB создаёт новый коннект к базе и заворачивает в контроллер
func ConnetctDB(dbUser, dbPass, dbHost, dbName string) *Error {
	newDB, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     dbHost,
			User:     dbUser,
			Password: dbPass,
			Port:     5432,
			Database: dbName,
		},
		MaxConnections: 50,
	})
	if err != nil {
		return NewError(InternalDatabase, err.Error())
	}

	db = newDB
	if err := Load(); err != nil {
		return err
	}

	return nil
}

// GetDB возвращает указатель на контроллер,
// для дополнительных запросов, не включенных в модели;
func GetDB() *pgx.ConnPool {
	return db
}
