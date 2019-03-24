package models

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

//easyjson:json
type Post struct {
	ID       int64     `json:"id"`
	Message  string    `json:"message"`
	IsEdited bool      `json:"isEdited"`
	Author   string    `json:"author"`
	Forum    string    `json:"forum"`
	Thread   int64     `json:"thread"`
	Parent   int64     `json:"parent"`
	Created  time.Time `json:"created"`

	parentImpl *int64
}

//easyjson:json
type Posts []*Post

//easyjson:json
type PostFull struct {
	Post   *Post   `json:"post"`
	Author *User   `json:"author"`
	Forum  *Forum  `json:"forum"`
	Thread *Thread `json:"thread"`
}

func (p *Post) Validate() *Error {
	if p.Message == "" {
		return NewError(ValidationFailed, "message is empty")
	}

	return nil
}

func (ps Posts) Create(slugOrID interface{}) *Error {
	tx, err := db.Begin()
	if err != nil {
		return NewError(InternalDatabase, "can not open posts create tx")
	}
	defer tx.Rollback()

	var thread *Thread
	switch v := slugOrID.(type) {
	case int64:
		id := strconv.FormatInt(v, 10)
		thread, _ = getThreadBy(tx, "id", &id)
	case string:
		thread, _ = getThreadBy(tx, "slug", &v)
	}
	if thread == nil {
		return NewError(ForeignKeyNotFound, "thread not found")
	}

	var createError *Error
	for _, p := range ps {
		createError = p.createImpl(tx, thread)
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

func (p *Post) createImpl(q queryer, thread *Thread) *Error {
	if validateError := p.Validate(); validateError != nil {
		return validateError
	}

	p.Forum = thread.Forum // можно убрать поле форум в табоице posts
	p.Thread = thread.ID

	if p.Parent != 0 {
		p.parentImpl = &p.Parent
		row := q.QueryRow(`SELECT p.thread FROM posts p WHERE p.id = $1;`, p.parentImpl)

		var parentThread int64
		if err := row.Scan(&parentThread); err != nil || parentThread != p.Thread {
			return NewError(ForeignKeyConflict, "Parent post was created in another thread")
		}
	}

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	// PostgreSQL считает с точностью до MS
	p.Created = p.Created.Round(time.Millisecond)
	newRow, err := q.Query(`INSERT INTO posts (message, created, author, forum, thread, parent, parents)
	VALUES ($1, $2, $3, $4, $5, $6, (SELECT parents FROM posts WHERE posts.id = $6) || (SELECT currval('posts_id_seq'))) RETURNING id`,
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

func (p *Post) Update() *Error {
	tx, err := db.Begin()
	if err != nil {
		return NewError(InternalDatabase, "can not open 'thread update' transaction")
	}
	defer tx.Rollback()

	storedPost := &Post{}
	row := tx.QueryRow(`SELECT p.id, p.message, p.is_edited, p.created, p.author, p.forum, p.thread, p.parent FROM posts p WHERE p.id = $1;`, p.ID)
	if err := row.Scan(&storedPost.ID, &storedPost.Message,
		&storedPost.IsEdited, &storedPost.Created, &storedPost.Author,
		&storedPost.Forum, &storedPost.Thread, &storedPost.parentImpl); err != nil {
		if err == sql.ErrNoRows {
			return NewError(RowNotFound, err.Error())
		}

		return NewError(InternalDatabase, err.Error())
	}
	if storedPost.parentImpl == nil {
		storedPost.Parent = 0
	} else {
		storedPost.Parent = *storedPost.parentImpl
	}

	needsUpdate := false
	if p.Message != "" && storedPost.Message != p.Message {
		needsUpdate = true
		storedPost.Message = p.Message
		storedPost.IsEdited = true
	}
	*p = *storedPost

	if !needsUpdate {
		return nil
	}

	_, err = tx.Exec(`UPDATE posts SET (message, is_edited) = ($1, $2) WHERE id = $3`, p.Message, p.IsEdited, p.ID)
	if err != nil {
		return NewError(InternalDatabase, err.Error())
	}

	if err = tx.Commit(); err != nil {
		return NewError(InternalDatabase, "thread update transaction commit error")
	}

	return nil
}

func GetPostByID(id int64, scope []string) (*PostFull, *Error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}
	defer tx.Rollback()

	pf := &PostFull{}
	var getErr *Error

	pf.Post, getErr = getPostByIDImpl(tx, id)
	if getErr != nil {
		return nil, getErr
	}

	for _, sc := range scope {
		switch sc {
		case "user":
			pf.Author, getErr = getUserBy(tx, "nickname", pf.Post.Author)
		case "forum":
			pf.Forum, getErr = getForumBySlugImpl(tx, pf.Post.Forum)
		case "thread":
			id := strconv.FormatInt(pf.Post.Thread, 10)
			pf.Thread, getErr = getThreadBy(tx, "id", &id)
		}
		if getErr != nil {
			return nil, getErr
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}
	return pf, nil
}

func getPostByIDImpl(q queryer, id int64) (*Post, *Error) {
	p := &Post{}

	row := q.QueryRow(`SELECT p.id, p.message, p.is_edited, p.created, p.author, p.forum, p.thread, p.parent FROM posts p WHERE p.id = $1;`, id)
	if err := row.Scan(&p.ID, &p.Message,
		&p.IsEdited, &p.Created, &p.Author,
		&p.Forum, &p.Thread, &p.parentImpl); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewError(RowNotFound, err.Error())
		}

		return nil, NewError(InternalDatabase, err.Error())
	}
	if p.parentImpl == nil {
		p.Parent = 0
	} else {
		p.Parent = *p.parentImpl
	}

	return p, nil
}

type SortMode int

const (
	Flat SortMode = iota
	Tree
	ParentTree
)

func GetPostsByThreadID(threadID int64, limit int, since int64, mode SortMode, desc bool) (Posts, *Error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, NewError(InternalDatabase, "get posts trans open error")
	}
	defer tx.Rollback()

	res, getError := getPostsByThreadIDImpl(tx, threadID, limit, since, mode, desc)
	if getError != nil {
		return nil, getError
	}

	err = tx.Commit()
	if err != nil {
		return nil, NewError(InternalDatabase, "get posts trans commit error")
	}

	return res, nil
}

func GetPostsByThreadSlug(slug string, limit int, since int64, mode SortMode, desc bool) (Posts, *Error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, NewError(InternalDatabase, "get posts trans open error")
	}
	defer tx.Rollback()

	thread, _ := getThreadBy(tx, "slug", &slug)
	if thread == nil {
		return nil, NewError(RowNotFound, "thread not exists")
	}

	res, getError := getPostsByThreadIDImpl(tx, thread.ID, limit, since, mode, desc)
	if getError != nil {
		return nil, getError
	}

	err = tx.Commit()
	if err != nil {
		return nil, NewError(InternalDatabase, "get posts trans commit error")
	}

	return res, nil
}

func getPostsByThreadIDImpl(q queryer, threadID int64, limit int, since int64, mode SortMode, desc bool) (Posts, *Error) {
	// АХТУНГ, страшный говнокод :(
	query := strings.Builder{}
	args := []interface{}{}
	switch mode {
	case Flat:
		args = append(args, threadID)
		query.WriteString(`SELECT p.id, p.message, p.is_edited, p.created, p.author,
						 p.forum, p.thread, p.parent FROM posts p WHERE p.thread = $1`)
		if since != 0 {
			args = append(args, since)
			query.WriteString(` AND (p.created, p.id) `)
			if desc {
				query.WriteByte('<')
			} else {
				query.WriteByte('>')
			}
			query.WriteString(` (SELECT posts.created, posts.id FROM posts WHERE posts.id=$2)`)
		}
		query.WriteString(` ORDER BY (p.created, p.id)`)
		if desc {
			query.WriteString(" DESC")
		}
		if limit != -1 {
			query.WriteString(" LIMIT $")
			query.WriteString(strconv.Itoa(len(args) + 1))
			args = append(args, limit)
		}
	case Tree:
		args = append(args, threadID)
		query.WriteString(`SELECT p.id, p.message, p.is_edited, p.created, p.author,
			p.forum, p.thread, p.parent FROM posts p WHERE p.thread = $1`)
		if since != 0 {
			args = append(args, since)
			query.WriteString(" AND p.parents ")
			if desc {
				query.WriteByte('<')
			} else {
				query.WriteByte('>')
			}
			query.WriteString(` (SELECT posts.parents FROM posts WHERE posts.id = $2)`)
		}
		query.WriteString(" ORDER BY p.parents")
		if desc {
			query.WriteString(" DESC")
		}
		if limit != -1 {
			query.WriteString(" LIMIT $")
			query.WriteString(strconv.Itoa(len(args) + 1))
			args = append(args, limit)
		}
	case ParentTree:
		args = append(args, threadID)
		query.WriteString(`SELECT p.id, p.message, p.is_edited, p.created, p.author,
			p.forum, p.thread, p.parent FROM posts p WHERE p.parents[1] IN (
				SELECT posts.id FROM posts WHERE posts.thread = $1 AND posts.parent IS NULL`)
		if since != 0 {
			args = append(args, since)
			query.WriteString(` AND posts.id`)
			if desc {
				query.WriteByte('<')
			} else {
				query.WriteByte('>')
			}
			query.WriteString(` (SELECT COALESCE(posts.parent, posts.id) FROM posts WHERE posts.id = $2)`)
		}
		query.WriteString(" ORDER BY posts.id")
		if desc {
			query.WriteString(" DESC")
		}
		if limit != -1 {
			query.WriteString(" LIMIT $")
			query.WriteString(strconv.Itoa(len(args) + 1))
			args = append(args, limit)
		}
		query.WriteString(`) ORDER BY`)
		if desc {
			query.WriteString(` p.parents[1] DESC,`)
		}
		query.WriteString(` p.parents`)
	}
	query.WriteByte(';')

	formatedID := strconv.FormatInt(threadID, 10)
	thread, _ := getThreadBy(q, "id", &formatedID)
	if thread == nil {
		return nil, NewError(RowNotFound, "no posts for this thread")
	}

	rows, err := q.Query(query.String(), args...)
	if err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}

	posts := make([]*Post, 0)
	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Message,
			&p.IsEdited, &p.Created, &p.Author,
			&p.Forum, &p.Thread, &p.parentImpl)
		if err != nil {
			return nil, NewError(InternalDatabase, err.Error())
		}
		if p.parentImpl == nil {
			p.Parent = 0
		} else {
			p.Parent = *p.parentImpl
		}

		posts = append(posts, p)
	}
	rows.Close()

	return posts, nil
}
