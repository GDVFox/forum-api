package models

import (
	"database/sql"
	"log"
	"regexp"
	"time"

	"github.com/lib/pq"
)

//easyjson:json
type Thread struct {
	ID      int64     `json:"id"`
	Slug    *string   `json:"slug"`
	Title   string    `json:"title"`
	Message string    `json:"message"`
	Votes   int32     `json:"votes"`
	Created time.Time `json:"created"`
	Author  string    `json:"author"`
	Forum   string    `json:"forum"`
}

var (
	threadSlugRegexp *regexp.Regexp
)

func init() {
	var err error
	threadSlugRegexp, err = regexp.Compile(`^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$`)
	if err != nil {
		log.Fatalf("slug regexp err: %s", err.Error())
	}
}

// Validate проверка полей
func (t *Thread) Validate() *Error {
	if !((t.Slug == nil || (t.Slug != nil && threadSlugRegexp.MatchString(*t.Slug))) &&
		t.Title != "" && t.Message != "") {
		return NewError(ValidationFailed, "validation failed")
	}

	return nil
}

func (t *Thread) Create() (*Thread, *Error) {
	if validateError := t.Validate(); validateError != nil {
		return nil, validateError
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, NewError(InternalDatabase, "can not open 'thread create' transaction")
	}
	defer tx.Rollback()

	duplicate, _ := getThreadDuplicate(tx, t.Slug, t.Author, t.Forum)
	if duplicate != nil {
		return duplicate, nil
	}

	newRow, err := tx.Query(`INSERT INTO threads (slug, title, message, created, author, forum)  VALUES ($1, $2,  $3, $4, $5, $6) RETURNING id`,
		t.Slug, t.Title, t.Message, t.Created, t.Author, t.Forum)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == "23502" {
			return nil, NewError(ForeignKeyNotFound, pgerr.Error())
		}

		return nil, NewError(InternalDatabase, err.Error())
	}
	if !newRow.Next() {
		return nil, NewError(InternalDatabase, "row does not created")
	}
	newRow.Scan(&t.ID)
	newRow.Close()

	if err = tx.Commit(); err != nil {
		return nil, NewError(InternalDatabase, "thread create transaction commit error")
	}

	return nil, nil
}

func getThreadDuplicate(q queryer, slug *string, author, forum string) (*Thread, *Error) {
	thread := &Thread{}
	row := q.QueryRow(`SELECT t.id, t.slug, t.title, t.message, t.votes, t.created,
						t.author, t.forum FROM threads t WHERE slug = $1 AND author = $2 AND forum = $3;`, slug, author, forum)
	if err := row.Scan(&thread.ID, &thread.Slug, &thread.Title,
		&thread.Title, &thread.Message, &thread.Votes,
		&thread.Created, &thread.Author, &thread.Forum); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewError(RowNotFound, "row does not found")
		}

		return nil, NewError(InternalDatabase, err.Error())
	}

	return thread, nil
}
