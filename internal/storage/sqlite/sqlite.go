package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/linemk/gRPC_auth/internal/domain/models"
	"github.com/linemk/gRPC_auth/internal/storage"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3" // Импортируем SQLite драйвер
)

type Storage struct {
	db *sql.DB // Поле для работы с базой данных через sql.DB
}

func NewStorage(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"
	// Открываем соединение с базой данных SQLite
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		// Возвращаем ошибку, если соединение не удалось открыть
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Возвращаем экземпляр Storage с открытой базой данных
	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error) {
	const op = "storage.sqlite.SaveUser"

	// Подготавливаем SQL-запрос для вставки пользователя
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO users (email, pass_hash) VALUES (?, ?)")
	if err != nil {
		// Возвращаем ошибку, если не удалось подготовить запрос
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Выполняем запрос с указанными значениями
	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		// Проверяем, если ошибка уникальности (например, email уже существует)
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			// Возвращаем ошибку существующего пользователя
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		// Возвращаем общую ошибку выполнения запроса
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Получаем ID последней вставленной записи
	id, err := res.LastInsertId()
	if err != nil {
		// Возвращаем ошибку, если не удалось получить ID
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	// Возвращаем ID нового пользователя
	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sqlite.User"

	// Подготавливаем SQL-запрос для выбора пользователя по email
	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email=?")
	if err != nil {
		// Возвращаем ошибку, если не удалось подготовить запрос
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	// Выполняем запрос с указанным email
	row := stmt.QueryRowContext(ctx, email)

	var user models.User

	// Читаем результат запроса в структуру пользователя
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		// Если пользователь не найден, возвращаем соответствующую ошибку
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		// Возвращаем другую ошибку, если произошел сбой
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	// Возвращаем найденного пользователя
	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"

	// Подготавливаем SQL-запрос для проверки, является ли пользователь админом
	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id=?")
	if err != nil {
		// Возвращаем ошибку, если не удалось подготовить запрос
		return false, fmt.Errorf("%s: %w", op, err)
	}

	// Выполняем запрос с указанным userID
	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool

	// Читаем результат запроса (флаг is_admin)
	err = row.Scan(&isAdmin)
	if err != nil {
		// Если пользователь не найден, возвращаем соответствующую ошибку
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		// Возвращаем другую ошибку, если произошел сбой
		return false, fmt.Errorf("%s: %w", op, err)
	}

	// Возвращаем результат (является ли пользователь администратором)
	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.sqlite.App"

	// Подготавливаем SQL-запрос для выбора приложения по ID
	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id=?")
	if err != nil {
		// Возвращаем ошибку, если не удалось подготовить запрос
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	// Выполняем запрос с указанным appID
	row := stmt.QueryRowContext(ctx, appID)

	var app models.App

	// Читаем результат запроса в структуру приложения
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		// Если приложение не найдено, возвращаем соответствующую ошибку
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		// Возвращаем другую ошибку, если произошел сбой
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	// Возвращаем найденное приложение
	return app, nil
}
