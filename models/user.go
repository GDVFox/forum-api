package models

// User информация о пользователе
type User struct {
	ID       int64  `json:"-"`
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname"`
	About    string `json:"about"`
	Email    string `json:"email"`
}

// GetUserByNickname получение информации о пользователе форума по его имени.
// пока не имеет смысла делать какой-то другой гет(так как все запросы через nickname)
func GetUserByNickname(nickname string) (*User, *Error) {
	rows, err := db.Query(`SELECT * FROM users WHERE nickname = $1`, nickname)
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
