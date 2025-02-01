package main

import (
	"github.com/linemk/gRPC_auth/internal/app"
	"github.com/linemk/gRPC_auth/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting application", slog.Any("env", cfg))
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCSrv.MustRun()

	// принудительная остановка
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSign := <-stop

	log.Info("received signal", slog.String("signal", stopSign.String()))
	application.GRPCSrv.Stop()
	log.Info("application stopped")
}
