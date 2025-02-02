package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/linemk/gRPC_auth/internal/domain/models"
	"github.com/linemk/gRPC_auth/internal/lib/jwt"
	"github.com/linemk/gRPC_auth/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger  // Логгер для записи информации, предупреждений и ошибок
	userSaver    UserSaver     // Интерфейс для сохранения пользователей
	userProvider UserProvider  // Интерфейс для получения данных пользователя
	appProvider  AppProvider   // Интерфейс для получения данных приложения
	tokenTTL     time.Duration // Время жизни токена в формате Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error) // Метод интерфейса для сохранения нового пользователя
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error) // Метод интерфейса для получения пользователя по email
	IsAdmin(ctx context.Context, userID int64) (bool, error)     // Метод интерфейса для проверки, является ли пользователь администратором
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error) // Метод интерфейса для получения данных о приложении по appID
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials") // Ошибка неверных учетных данных
	ErrInvalidAppID       = errors.New("invalid app id")      // Ошибка некорректного идентификатора приложения
	ErrUserExists         = errors.New("user already exists") // Ошибка, если пользователь уже существует
)

// New создает объект Auth для работы сервиса
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,          // Устанавливает логгер
		userSaver:    userSaver,    // Устанавливает объект для сохранения пользователей
		userProvider: userProvider, // Устанавливает объект для получения информации о пользователях
		appProvider:  appProvider,  // Устанавливает объект для получения информации о приложениях
		tokenTTL:     tokenTTL,     // Устанавливает время жизни токена
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "auth.Login" // Название операции для логирования

	log := a.log.With(
		slog.String("op", op),          // Добавляет название операции в лог
		slog.String("username", email), // Добавляет email пользователя в лог
	)
	log.Info("checking user")                    // Логирует, что начата проверка пользователя
	user, err := a.userProvider.User(ctx, email) // Получение информации о пользователе по email
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) { // Если пользователь не найден
			a.log.Warn("user not found", slog.String("email", email))  // Логирует предупреждение о том, что пользователь не найден
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials) // Возвращает ошибку "неверные учетные данные"
		}
		a.log.Error("failed to get user", err)   // Логирует ошибку получения пользователя
		return "", fmt.Errorf("%s: %w", op, err) // Возвращает ошибку
	}
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil { // Сравнивает хэш пароля с предоставленным паролем
		a.log.Warn("invalid password", slog.String("email", email)) // Логирует предупреждение о некорректном пароле
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)  // Возвращает ошибку "неверные учетные данные"
	}
	app, err := a.appProvider.App(ctx, appID) // Получает данные приложения по appID
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err) // Возвращает ошибку при получении приложения
	}
	log.Info("user logged in", slog.String("email", email)) // Логирует успешную авторизацию пользователя

	token, err := jwt.NewToken(user, app, a.tokenTTL) // Генерирует новый JWT токен для пользователя
	if err != nil {
		a.log.Error("failed to generate token", err) // Логирует ошибку генерации токена
		return "", fmt.Errorf("%s: %w", op, err)     // Возвращает ошибку
	}
	return token, nil // Возвращает токен
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.RegisterNewUser" // Название операции для логирования

	log := a.log.With(
		slog.String("op", op),       // Добавляет название операции в лог
		slog.String("email", email), // Добавляет email пользователя в лог
	)

	log.Info("registreting new user")                                                  // Логирует начало регистрации нового пользователя
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // Генерирует хэш для указанного пароля
	if err != nil {
		log.Error("failed to hash password", err) // Логирует ошибку при создании хэша пароля
		return 0, fmt.Errorf("%s: %w", op, err)   // Возвращает ошибку
	}

	// сохраняем в БД
	id, err := a.userSaver.SaveUser(ctx, email, passHash) // Сохраняет нового пользователя в БД
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) { // Если пользователь уже существует
			log.Warn("user already exists", err)              // Логирует предупреждение о существующем пользователе
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists) // Возвращает ошибку "пользователь уже существует"
		}
		log.Error("failed to save user", err)   // Логирует ошибку сохранения пользователя
		return 0, fmt.Errorf("%s: %w", op, err) // Возвращает ошибку
	}

	log.Info("user created", slog.String("email", email)) // Логирует успешное создание пользователя
	return id, nil                                        // Возвращает идентификатор пользователя
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin" // Название операции для логирования
	log := a.log.With(
		slog.String("op", op),         // Добавляет название операции в лог
		slog.Int64("user_id", userID), // Добавляет идентификатор пользователя в лог
	)
	log.Info("checking if user is Admin")               // Логирует начало проверки, является ли пользователь администратором
	isAdmin, err := a.userProvider.IsAdmin(ctx, userID) // Проверяет, является ли пользователь администратором
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) { // Если приложение не найдено
			log.Warn("app not found", err)                          // Логирует предупреждение, что приложение не найдено
			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID) // Возвращает ошибку
		}
		return false, fmt.Errorf("%s: %w", op, err) // Возвращает ошибку
	}
	log.Info("checked if user is Admin", slog.Bool("Is_Admin", isAdmin)) // Логирует результат проверки
	return isAdmin, nil                                                  // Возвращает результат проверки
}
