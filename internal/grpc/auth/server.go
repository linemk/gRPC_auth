package auth

import (
	"context"
	"errors"
	"github.com/linemk/gRPC_auth/internal/services/auth" // Импортируем сервисы для авторизации
	"github.com/linemk/gRPC_auth/internal/storage"       // Импортируем хранилище
	ssov1 "github.com/linemk/proto_buf/gen/go/sso"       // Импортируем сгенерированные protobuf файлы
	"google.golang.org/grpc"                             // Импортируем gRPC библиотеку
	"google.golang.org/grpc/codes"                       // Импортируем коды статусов gRPC
	"google.golang.org/grpc/status"                      // Импортируем статус gRPC
)

// Интерфейс для работы с авторизацией
type Auth interface {
	// Метод входа пользователя с получением токена
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	// Метод регистрации нового пользователя
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	// Метод проверки, является ли пользователь администратором
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

// gRPC сервер для работы с API авторизации
type ServerApi struct {
	ssov1.UnimplementedAuthServer      // Встраиваем несгенерированные методы сервера
	auth                          Auth // Включаем интерфейс для авторизации
}

// Регистрируем сервис авторизации на gRPC сервере
func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerApi{auth: auth}) // Регистрируем AuthServer на gRPC
}

const (
	emptyValue = 0 // Константа для обозначения пустого значения
)

// Метод обработки авторизации зарегистрированных пользователей
func (s *ServerApi) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil { // Валидируем запрос
		return nil, err // Возвращаем ошибку, если валидация не прошла
	}
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId())) // Пытаемся залогинить пользователя
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) { // Проверяем, является ли ошибка ошибкой неверных данных
			return nil, status.Error(codes.Unauthenticated, err.Error()) // Возвращаем ошибку авторизации
		}
		return nil, status.Error(codes.Internal, "internal server error") // Возвращаем внутреннюю ошибку
	}

	return &ssov1.LoginResponse{
		Token: token, // Возвращаем сгенерированный токен
	}, nil
}

// Метод обработки регистрации новых пользователей
func (s *ServerApi) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil { // Валидируем запрос
		return nil, err // Возвращаем ошибку, если валидация не прошла
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword()) // Пытаемся зарегистрировать нового пользователя
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) { // Проверяем, является ли ошибка ошибкой о существовании пользователя
			return nil, status.Error(codes.AlreadyExists, "user already exists") // Возвращаем ошибку существующего пользователя
		}
		return nil, status.Error(codes.Internal, "internal server error") // Возвращаем внутреннюю ошибку
	}
	return &ssov1.RegisterResponse{
		UserId: userID, // Возвращаем идентификатор пользователя
	}, nil
}

// Метод проверки, является ли пользователь администратором
func (s *ServerApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil { // Валидируем запрос
		return nil, err // Возвращаем ошибку, если валидация не прошла
	}

	userID, err := s.auth.IsAdmin(ctx, req.GetUserId()) // Проверяем, является ли пользователь администратором
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) { // Проверяем, является ли ошибка ошибкой отсутствия пользователя
			return nil, status.Error(codes.NotFound, "user not found") // Возвращаем ошибку, если пользователь не найден
		}

		return nil, status.Error(codes.Internal, "internal server error") // Возвращаем внутреннюю ошибку
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: userID, // Возвращаем информацию, является ли пользователь администратором
	}, nil
}

// Валидатор для входа пользователя
func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" || req.GetPassword() == "" { // Проверяем, заполнены ли email и пароль
		return status.Error(codes.InvalidArgument, "Email or password is required") // Возвращаем ошибку, если они пусты
	}

	if req.GetAppId() == emptyValue { // Проверяем, заполнен ли AppId
		return status.Error(codes.InvalidArgument, "AppId is required") // Возвращаем ошибку, если он пуст
	}

	return nil // Возвращаем nil при успешной валидации
}

// Валидатор для регистрации нового пользователя
func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" || req.GetPassword() == "" { // Проверяем, заполнены ли email и пароль
		return status.Error(codes.InvalidArgument, "Email or password is required") // Возвращаем ошибку, если они пусты
	}

	return nil // Возвращаем nil при успешной валидации
}

// Валидатор для проверки, является ли пользователь администратором
func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.UserId == emptyValue { // Проверяем, заполнен ли UserId
		return status.Error(codes.InvalidArgument, "AppId is required") // Возвращаем ошибку, если он пуст
	}

	return nil // Возвращаем nil при успешной валидации
}
