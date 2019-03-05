package models

import (
	"log"
	"regexp"

	"github.com/lib/pq"
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
	slugRegexp *regexp.Regexp
)

func init() {
	var err error
	slugRegexp, err = regexp.Compile(`^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$`)
	if err != nil {
		log.Fatalf("slug regexp err: %s", err.Error())
	}
}

// Validate проверка полей
func (f *Forum) Validate() *Error {
	if !(slugRegexp.MatchString(f.Slug) &&
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

	usedSlug, _ := GetForumBySlug(f.Slug)
	if usedSlug != nil {
		return usedSlug, NewError(RowDuplication, "slug is already used!")
	}

	// слегка странное решение, чтобы обеспечить сопадение nickname в таблице users и forums,
	// а также сохранить регистронезависимость;
	newRow, err := db.Query(`INSERT INTO forums (slug, title, owner) VALUES ($1, $2, (SELECT nickname FROM users WHERE nickname = $3)) RETURNING id, owner`,
		f.Slug, f.Title, f.Owner)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == "23502" {
			return nil, NewError(ForeignKeyNotFound, pgerr.Error())
		}

		return nil, NewError(InternalDatabase, err.Error())
	}
	defer newRow.Close()
	if !newRow.Next() {
		return nil, NewError(InternalDatabase, "row does not created")
	}
	// обновляем структуру, чтобы она содержала валидное имя создателя(учитывая регистр)
	// и валидный ID
	newRow.Scan(&f.ID, &f.Owner)

	return nil, nil
}

// GetForumBySlug получение информации о форуме по его идентификатору
func GetForumBySlug(slug string) (*Forum, *Error) {
	rows, err := db.Query(`SELECT f.id, f.slug, f.title, f.posts, f.threads, f.owner FROM forums f WHERE slug = $1`, slug)
	if err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, NewError(RowNotFound, "row does not found")
	}

	forum := &Forum{}
	rows.Scan(&forum.ID, &forum.Slug, &forum.Title,
		&forum.Posts, &forum.Threads, &forum.Owner)

	return forum, nil
}
