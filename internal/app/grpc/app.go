package grpcapp

import (
	"fmt"

	authgrpc "github.com/linemk/gRPC_auth/internal/grpc/auth" // Пакет для работы с gRPC авторизацией
	"google.golang.org/grpc"                                  // gRPC библиотека
	"log/slog"                                                // Логирование
	"net"                                                     // Работа с сетевыми соединениями
)

// App представляет gRPC-приложение
type App struct {
	log        *slog.Logger // Логгер для записи событий
	gRPCServer *grpc.Server // gRPC сервер
	port       int          // Порт, на котором запускается сервер
}

// New создает новый экземпляр App
func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	gRPCServer := grpc.NewServer()             // Создаем новый gRPC сервер
	authgrpc.Register(gRPCServer, authService) // Регистрируем сервис авторизации в gRPC сервере
	return &App{
		log:        log,        // Устанавливаем логгер
		gRPCServer: gRPCServer, // Устанавливаем gRPC сервер
		port:       port,       // Устанавливаем порт сервера
	}
}

// MustRun запускает сервер и паникует при ошибке
func (a *App) MustRun() {
	if err := a.Run(); err != nil { // Запускаем сервер
		panic(err) // Если возникает ошибка, падаем с паникой
	}
}

// Run запускает gRPC сервер
func (a *App) Run() error {
	const op = "grpc.Run"                                              // Обозначаем операцию для логирования
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port)) // Добавляем в лог информацию о порте и операции

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port)) // Открываем TCP соединение на заданном порту
	if err != nil {
		return fmt.Errorf("%s:%w", op, err) // Возвращаем ошибку при невозможности открыть соединение
	}
	log.Info(" grpc server is running", slog.String("addr", l.Addr().String())) // Логируем успешный запуск gRPC сервера
	if err := a.gRPCServer.Serve(l); err != nil {                               // Запускаем gRPC сервер на прослушивании соединения
		return fmt.Errorf("%s:%w", op, err) // Возвращаем ошибку в случае сбоя
	}
	return nil // Возвращаем nil при успешном запуске
}

// Stop останавливает gRPC сервер
func (a *App) Stop() {
	const op = "grpc.Stop"                                     // Обозначаем операцию для логирования
	log := a.log.With(slog.String("op", op))                   // Добавляем в лог информацию об операции
	log.Info("stopping grpc server", slog.Int("port", a.port)) // Логируем остановку сервера
	a.gRPCServer.GracefulStop()                                // Останавливаем сервер с завершением активных соединений
}
