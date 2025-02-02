package models

type User struct {
	ID       int64  // Уникальный идентификатор пользователя.
	Email    string // Электронная почта пользователя.
	PassHash []byte // Хэш пароля пользователя.
}
