package models

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"

	"github.com/lib/pq"
)

// User информация о пользователе
//easyjson:json
type User struct {
	ID       int64  `json:"-"`
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
}

// UpdateUserFields структура для обновления полей юзера
//easyjson:json
type UpdateUserFields struct {
	Fullname *string `json:"fullname"`
	About    *string `json:"about"`
	Email    *string `json:"email"`
}

// Users несколько юзеров
//easyjson:json
type Users []*User

var (
	nicknameRegexp *regexp.Regexp
	emailRegexp    *regexp.Regexp
)

func init() {
	var err error
	nicknameRegexp, err = regexp.Compile(`^[a-zA-Z0-9_.]+$`)
	if err != nil {
		log.Fatalf("nickname regexp err: %s", err.Error())
	}

	emailRegexp, err = regexp.Compile("^[a-zA-Z0-9.!#$%&''*+/=?^_`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`)
	if err != nil {
		log.Fatalf("email regexp err: %s", err.Error())
	}
}

// Validate проверка полей
func (u *User) Validate() *Error {
	if !(nicknameRegexp.MatchString(u.Nickname) &&
		emailRegexp.MatchString(u.Email) &&
		u.Fullname != "") {
		return NewError(ValidationFailed, "validation failed")
	}

	return nil
}

// Create создание нового пользователя в базе данных
func (u *User) Create() (Users, *Error) {
	if validateError := u.Validate(); validateError != nil {
		return nil, validateError
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, NewError(InternalDatabase, "can not open 'user create' transaction")
	}
	defer tx.Rollback()

	// валдация на повторы
	usedUsers := getDuplicates(tx, u.Nickname, u.Email)
	if usedUsers != nil {
		return usedUsers, NewError(RowDuplication, "email or nickname are already used!")
	}

	newRow, err := tx.Query(`INSERT INTO users (nickname, fullname, about, email) VALUES ($1, $2, $3, $4) RETURNING id`,
		u.Nickname, u.Fullname, u.About, u.Email)
	if err != nil {
		return nil, NewError(InternalDatabase, err.Error())
	}
	if !newRow.Next() {
		return nil, NewError(RowNotFound, "row does not found")
	}

	// обновляем структуру так, чтобы она содержала валидный id
	newRow.Scan(&u.ID)
	newRow.Close()

	err = tx.Commit()
	if err != nil {
		return nil, NewError(InternalDatabase, "user create transaction commit error")
	}

	return nil, nil
}

// Save сохраняет user с новыми полями
func (u *User) Save() *Error {
	if u.ID == 0 {
		return NewError(ValidationFailed, "ID must be setted")
	}

	if err := u.Validate(); err != nil {
		return err
	}

	// возможно далее указывать в запросе не все поля
	_, err := db.Exec(`UPDATE users SET (nickname, fullname, about, email) = ($1, $2, $3, $4) WHERE id = $5`,
		u.Nickname, u.Fullname, u.About, u.Email, u.ID)
	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok && pgerr.Code == "23505" {
			return NewError(RowDuplication, pgerr.Error())
		}

		return NewError(InternalDatabase, err.Error())
	}

	return nil
}

// GetUserByNickname получение информации о пользователе форума по егsо имени.
func GetUserByNickname(nickname string) (*User, *Error) {
	return getUserBy(db, "nickname", nickname)
}

// GetUserByEmail получение информации о пользователе форума по егsо email.
func GetUserByEmail(email string) (*User, *Error) {
	return getUserBy(db, "email", email)
}

func getDuplicates(q queryer, nickname, email string) Users {
	usedUsers := make([]*User, 0)

	dupNickname, _ := getUserBy(q, "nickname", nickname)
	if dupNickname != nil {
		usedUsers = append(usedUsers, dupNickname)
	}

	dupEmail, _ := getUserBy(q, "email", email)
	if dupEmail != nil && (dupNickname == nil || dupEmail.ID != dupNickname.ID) {
		usedUsers = append(usedUsers, dupEmail)
	}

	if len(usedUsers) == 0 {
		return nil
	}

	return usedUsers
}

func getUserBy(q queryer, by, value string) (*User, *Error) {
	user := &User{}
	// спринтф затратно, потом надо это ускорить
	row := q.QueryRow(fmt.Sprintf(`SELECT * FROM users WHERE %s = $1`, by), value)
	if err := row.Scan(&user.ID, &user.Nickname,
		&user.Fullname, &user.About, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewError(RowNotFound, "row does not found")
		}

		return nil, NewError(InternalDatabase, err.Error())
	}

	return user, nil
}
