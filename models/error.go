package models

// Error объект для удобства обработки ошибок
type Error struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

// NewError создаёт новый объект ошибки
func NewError(code int, msg string) *Error {
	return &Error{
		Code:    code,
		Message: msg,
	}
}

const (
	// InternalDatabase неизвестная ошибка базы данных
	InternalDatabase = 700 + iota
	// RowNotFound запись в БД не найдена
	RowNotFound
	// ValidationFailed объект не прошел валидацию
	ValidationFailed
	// RowDuplication юзер с таким именем или почтой уже существует
	RowDuplication
	// ForeignKeyNotFound запись на которую ссылается не найдена
	ForeignKeyNotFound
	// ForeignKeyConflict запись на которую ссылаемся некорректна(например родитель в другом треде)
	ForeignKeyConflict
)
