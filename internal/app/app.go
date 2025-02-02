package app

import (
	grpcapp "github.com/linemk/gRPC_auth/internal/app/grpc" // Импорт модуля gRPC приложения
	"github.com/linemk/gRPC_auth/internal/services/auth"    // Импорт модуля сервиса авторизации
	"github.com/linemk/gRPC_auth/internal/storage/sqlite"   // Импорт модуля хранилища, реализованного на SQLite

	"log/slog" // Импорт логгера
	"time"     // Импорт пакета для работы со временем
)

// App представляет основное приложение
type App struct {
	GRPCSrv *grpcapp.App // gRPC сервер приложения
}

// New создает новый экземпляр App
func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	storage, err := sqlite.NewStorage(storagePath) // Инициализируем SQLite хранилище
	if err != nil {
		panic(err) // Завершаем работу приложения, если хранилище не удалось инициализировать
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL) // Создаем сервис авторизации

	grpcApp := grpcapp.New(log, authService, grpcPort) // Создаем gRPC приложение
	return &App{
		GRPCSrv: grpcApp, // Записываем gRPC сервер в основное приложение
	}
}
