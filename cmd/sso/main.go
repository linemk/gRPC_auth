package main

import (
	"github.com/linemk/gRPC_auth/internal/app"    // Импорт приложения
	"github.com/linemk/gRPC_auth/internal/config" // Импорт загрузчика конфигурации
	"log/slog"                                    // Импорт библиотеки логирования
	"os"                                          // Импорт для работы с ОС
	"os/signal"                                   // Импорт для обработки сигналов ОС
	"syscall"                                     // Импорт системных вызовов
)

const (
	envLocal = "local" // Среда разработки - локальная
	envDev   = "dev"   // Среда разработки - dev
	envProd  = "prod"  // Среда разработки - production
)

// setupLogger настраивает логгер в зависимости от среды
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger // Переменная для логгера

	switch env { // Логика выбора обработчика логов
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}), // Логгер для local: текстовый вывод с уровнем отладки
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}), // Логгер для dev: JSON вывод с уровнем отладки
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}), // Логгер для prod: JSON вывод с уровнем информации
		)
	}
	return log // Возвращаем настроенный логгер
}

func main() {
	cfg := config.MustLoad() // Загружаем конфигурацию

	log := setupLogger(cfg.Env)                            // Настраиваем логгер в зависимости от среды
	log.Info("starting application", slog.Any("env", cfg)) // Логгируем запуск приложения

	// Создаем объект приложения
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCSrv.MustRun() // Запускаем gRPC сервер в отдельной го-рутине

	// Создаем канал для получения сигнала остановки
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM) // Подписываемся на сигналы завершения из ОС
	stopSign := <-stop                                   // Ожидаем сигнал остановки

	log.Info("received signal", slog.String("signal", stopSign.String())) // Логгируем полученный сигнал
	application.GRPCSrv.Stop()                                            // Останавливаем gRPC сервер
	log.Info("application stopped")                                       // Логгируем остановку приложения
}
