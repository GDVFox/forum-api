package models

import (
	"fmt"
	"log"
	"regexp"
)

// User информация о пользователе
type User struct {
	ID       int64  `json:"-"`
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
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
	nicknameRegexp, err = regexp.Compile(`^[a-zA-Z0-9_]+$`)
	if err != nil {
		log.Fatalf("nickname regexp err: %s", err.Error())
	}

	emailRegexp, err = regexp.Compile("^[a-zA-Z0-9.!#$%&''*+/=?^_`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`)
	if err != nil {
		log.Fatalf("email regexp err: %s", err.Error())
	}
}

// Validate проверка полей
func (u *User) Validate() bool {
	return nicknameRegexp.MatchString(u.Nickname) &&
		emailRegexp.MatchString(u.Email) &&
		u.Fullname != ""
}

// Create создание нового пользователя в базе данных
func (u *User) Create() (Users, *Error) {
	if !u.Validate() {
		return nil, NewError(ValidationFailed, "validation failed", "")
	}

	usedUsers := u.getDuplicates()
	if usedUsers != nil {
		return usedUsers, NewError(RowDuplication, "email or nickname are already used!", "")
	}

	_, err := db.Exec(`INSERT INTO users (nickname, fullname, about, email) VALUES ($1, $2, $3, $4)`,
		u.Nickname, u.Fullname, u.About, u.Email)

	if err != nil {
		return nil, NewError(InternalDatabase, err.Error(), "")
	}

	return nil, nil
}

func (u *User) getDuplicates() Users {
	usedUsers := make([]*User, 0)

	dupNickname, _ := GetUserByNickname(u.Nickname)
	if dupNickname != nil {
		usedUsers = append(usedUsers, dupNickname)
	}

	dupEmail, _ := GetUserByEmail(u.Email)
	if dupEmail != nil && dupEmail.ID != dupNickname.ID {
		usedUsers = append(usedUsers, dupEmail)
	}

	if len(usedUsers) == 0 {
		return nil
	}

	return usedUsers
}

func getUserBy(by, value string) (*User, *Error) {
	// спринтф затратно, потом надо это ускорить
	rows, err := db.Query(fmt.Sprintf(`SELECT * FROM users WHERE %s = $1`, by), value)
	if err != nil {
		return nil, NewError(InternalDatabase, err.Error(), "")
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, NewError(RowNotFound, "row does not found", "")
	}

	// пока что самый простой подход,
	// чтобы не загружаться дополнительными уровнем рефлексии
	user := &User{}
	rows.Scan(&user.ID, &user.Nickname,
		&user.Fullname, &user.About, &user.Email)

	return user, nil
}

// GetUserByNickname получение информации о пользователе форума по егsо имени.
func GetUserByNickname(nickname string) (*User, *Error) {
	return getUserBy("nickname", nickname)
}

// GetUserByEmail получение информации о пользователе форума по егsо email.
func GetUserByEmail(email string) (*User, *Error) {
	return getUserBy("email", email)
}
