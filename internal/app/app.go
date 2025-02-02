package app

import (
	grpcapp "github.com/linemk/gRPC_auth/internal/app/grpc"
	"github.com/linemk/gRPC_auth/internal/services/auth"
	"github.com/linemk/gRPC_auth/internal/storage/sqlite"

	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	storage, err := sqlite.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)
	return &App{
		GRPCSrv: grpcApp,
	}
}
