package models

import (
	"strconv"
	"time"

	"github.com/lib/pq"
)

//easyjson:json
type Post struct {
	ID         int64     `json:"id"`
	Message    string    `json:"message"`
	IsEdited   bool      `json:"is_edited"`
	Author     string    `json:"author"`
	Forum      string    `json:"forum"`
	Thread     int64     `json:"thread"`
	ThreadSlug *string   `json:"-"`
	Parent     int64     `json:"parent"`
	Created    time.Time `json:"created"`

	parentImpl *int64
}

//easyjson:json
type Posts []*Post

func (p *Post) Validate() *Error {
	if p.Message == "" {
		return NewError(ValidationFailed, "message is empty")
	}

	return nil
}

func (p *Post) Create() *Error {
	return p.createImpl(db)
}

func (ps Posts) Create() *Error {
	tx, err := db.Begin()
	if err != nil {
		return NewError(InternalDatabase, "can not open posts create tx")
	}
	defer tx.Rollback()

	var createError *Error
	for _, p := range ps {
		createError = p.createImpl(tx)
		if createError != nil {
			return createError
		}
	}

	err = tx.Commit()
	if err != nil {
		return NewError(InternalDatabase, "con not commit posts create tx")
	}

	return nil
}

func (p *Post) createImpl(q queryer) *Error {
	if validateError := p.Validate(); validateError != nil {
		return validateError
	}

	var thread *Thread
	if p.ThreadSlug != nil {
		thread, _ = getThreadBy(q, "slug", p.ThreadSlug)
	} else {
		id := strconv.FormatInt(p.Thread, 10)
		thread, _ = getThreadBy(q, "id", &id)
	}
	if thread == nil {
		return NewError(RowNotFound, "thread not found")
	}

	p.Forum = thread.Forum // можно убрать поле форум в табоице posts
	p.Thread = thread.ID

	if p.Parent != 0 {
		p.parentImpl = &p.Parent
	}

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	newRow, err := q.Query(`INSERT INTO posts (message, created, author, forum, thread, parent) VALUES ($1, $2 AT TIME ZONE 'UTC', $3, $4, $5, $6) RETURNING id`,
		p.Message, p.Created, p.Author, p.Forum, p.Thread, p.parentImpl)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == "23503" {
			return NewError(ForeignKeyNotFound, pgerr.Error())
		}

		return NewError(InternalDatabase, err.Error())
	}
	if !newRow.Next() {
		return NewError(InternalDatabase, "row does not created")
	}
	// обновляем структуру, чтобы она содержала валидное имя создателя(учитывая регистр)
	// и валидный ID
	newRow.Scan(&p.ID)
	newRow.Close()

	return nil
}
