package models

// Error объект для удобства обработки ошибок
type Error struct {
	Code        int    `json:"-"`
	Message     string `json:"message"`
	Description string `json:"-"`
}

// NewError создаёт новый объект ошибки
func NewError(code int, descr, msg string) *Error {
	return &Error{
		Code:        code,
		Description: descr,
		Message:     msg,
	}
}

const (
	// InternalDatabase неизвестная ошибка базы данных
	InternalDatabase = 100 + iota
	// RowNotFound запись в БД не найдена
	RowNotFound
)
