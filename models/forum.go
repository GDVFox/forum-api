package models

import (
	"log"
	"regexp"

	"github.com/jackc/pgx"
)

// Forum информация о форуме.
//easyjson:json
type Forum struct {
	ID      int64  `json:"-"`
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Posts   int64  `json:"posts"`
	Threads int32  `json:"threads"`
	Owner   string `json:"user"`
}

var (
	forumSlugRegexp *regexp.Regexp
)

func init() {
	var err error
	forumSlugRegexp, err = regexp.Compile(`^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$`)
	if err != nil {
		log.Fatalf("slug regexp err: %s", err.Error())
	}
}

// Validate проверка полей
func (f *Forum) Validate() *Error {
	if !(forumSlugRegexp.MatchString(f.Slug) &&
		f.Title != "") {
		return NewError(ValidationFailed, "validation failed")
	}

	return nil
}

// Create cоздание нового форума.
func (f *Forum) Create() (*Forum, *Error) {
	if validateError := f.Validate(); validateError != nil {
		return nil, validateError
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, NewError(InternalDatabase, "can not open 'forum create' transaction")
	}
	defer tx.Rollback()

	usedSlug, _ := getForumBySlugImpl(tx, f.Slug)
	if usedSlug != nil {
		return usedSlug, NewError(RowDuplication, "slug is already used!")
	}

	// слегка странное решение, чтобы обеспечить сопадение nickname в таблице users и forums,
	// а также сохранить регистронезависимость;
	newRow, err := tx.Query(`INSERT INTO forums (slug, title, owner) VALUES ($1, $2, (SELECT nickname FROM users WHERE nickname = $3)) RETURNING id, owner`,
		f.Slug, f.Title, f.Owner)
	if err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}
	if !newRow.Next() {
		if pgerr, ok := newRow.Err().(pgx.PgError); ok && pgerr.Code == "23502" {
			return nil, NewError(ForeignKeyNotFound, pgerr.Error())
		}

		return nil, NewError(InternalDatabase, newRow.Err().Error())
	}
	// обновляем структуру, чтобы она содержала валидное имя создателя(учитывая регистр)
	// и валидный ID
	if err = newRow.Scan(&f.ID, &f.Owner); err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}
	newRow.Close()

	if err = tx.Commit(); err != nil {
		return nil, NewError(InternalDatabase, "forum create transaction commit error")
	}

	return nil, nil
}

// GetForumBySlug получение информации о форуме по его идентификатору
func GetForumBySlug(slug string) (*Forum, *Error) {
	return getForumBySlugImpl(db, slug)
}

func getForumBySlugImpl(q queryer, slug string) (*Forum, *Error) {
	forum := &Forum{}
	row := q.QueryRow(`SELECT f.id, f.slug, f.title, f.posts, f.threads, f.owner FROM forums f WHERE slug = $1`, slug)
	if err := row.Scan(&forum.ID, &forum.Slug, &forum.Title,
		&forum.Posts, &forum.Threads, &forum.Owner); err != nil {
		if err == pgx.ErrNoRows {
			return nil, NewError(RowNotFound, "row does not found")
		}

		return nil, NewError(InternalDatabase, err.Error())
	}

	return forum, nil
}
