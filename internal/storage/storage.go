package storage

import (
	"errors" // Пакет для работы с ошибками
)

var (
	ErrUserExists   = errors.New("user already exists") // Ошибка: пользователь уже существует
	ErrUserNotFound = errors.New("user not found")      // Ошибка: пользователь не найден
	ErrAppNotFound  = errors.New("app not found")       // Ошибка: приложение не найдено
)
