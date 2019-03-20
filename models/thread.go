package models

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

//easyjson:json
type Thread struct {
	ID      int64      `json:"id"`
	Slug    *string    `json:"slug"`
	Title   string     `json:"title"`
	Message string     `json:"message"`
	Votes   int32      `json:"votes"`
	Created *time.Time `json:"created"`
	Author  string     `json:"author"`
	Forum   string     `json:"forum"`
}

//easyjson:json
type Threads []*Thread

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

	duplicate, _ := getThreadBy(tx, "slug", t.Slug)
	if duplicate != nil {
		return duplicate, nil
	}

	newRow, err := tx.Query(`INSERT INTO threads (slug, title, message, created, author, forum)  VALUES ($1, $2, $3, $4 AT TIME ZONE 'UTC', $5, (SELECT slug FROM forums WHERE slug = $6)) RETURNING id, forum`,
		t.Slug, t.Title, t.Message, t.Created, t.Author, t.Forum)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && (pgerr.Code == "23503" || pgerr.Code == "23502") {
			return nil, NewError(ForeignKeyNotFound, pgerr.Error())
		}

		return nil, NewError(InternalDatabase, err.Error())
	}
	if !newRow.Next() {
		return nil, NewError(InternalDatabase, "row does not created")
	}
	newRow.Scan(&t.ID, &t.Forum)
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
	if err := row.Scan(&thread.ID, &thread.Slug,
		&thread.Title, &thread.Message, &thread.Votes,
		&thread.Created, &thread.Author, &thread.Forum); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewError(RowNotFound, "row does not found")
		}

		return nil, NewError(InternalDatabase, err.Error())
	}

	return thread, nil
}

func GetThreadsByForum(forumSlug string, limit int, since time.Time, desc bool) (Threads, *Error) {
	query := strings.Builder{}
	args := []interface{}{forumSlug}
	query.WriteString(`SELECT t.id, t.slug, t.title, t.message, t.votes, t.created,
						t.author, t.forum FROM threads t WHERE forum = $1`)
	if !since.IsZero() {
		query.WriteString(" AND created")
		if desc {
			query.WriteString(" <=")
		} else {
			query.WriteString(" >=")
		}
		query.WriteString(" $2")
		args = append(args, since)
	}
	query.WriteString(" ORDER BY t.created")
	if desc {
		query.WriteString(" DESC")
	}
	if limit != -1 {
		query.WriteString(" LIMIT $")
		query.WriteString(strconv.Itoa(len(args) + 1))
		args = append(args, limit)
	}
	query.WriteByte(';')

	tx, err := db.Begin()
	if err != nil {
		return nil, NewError(InternalDatabase, "can not open 'thread get' transaction")
	}
	defer tx.Rollback()

	forum, _ := getForumBySlugImpl(tx, forumSlug)
	if forum == nil {
		return nil, NewError(RowNotFound, "no threads for this forum")
	}
	rows, err := tx.Query(query.String(), args...)
	if err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}

	threads := make([]*Thread, 0)
	for rows.Next() {
		t := &Thread{}
		err = rows.Scan(&t.ID, &t.Slug,
			&t.Title, &t.Message, &t.Votes,
			&t.Created, &t.Author, &t.Forum)
		if err != nil {
			return nil, NewError(InternalDatabase, err.Error())
		}
		threads = append(threads, t)
	}
	rows.Close()

	if err = tx.Commit(); err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}

	return threads, nil
}

func getThreadBy(q queryer, by string, value *string) (*Thread, *Error) {
	t := &Thread{}
	// спринтф затратно, потом надо это ускорить
	row := q.QueryRow(fmt.Sprintf(`SELECT t.id, t.slug, t.title, t.message, t.votes, t.created,
									t.author, t.forum FROM threads t WHERE %s = $1`, by), value)
	if err := row.Scan(&t.ID, &t.Slug,
		&t.Title, &t.Message, &t.Votes,
		&t.Created, &t.Author, &t.Forum); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewError(RowNotFound, "row does not found")
		}

		return nil, NewError(InternalDatabase, err.Error())
	}

	return t, nil
}
