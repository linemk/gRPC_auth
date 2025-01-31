package app

import (
	"google.golang.org/grpc"
	"log/slog"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       string
}

func New(log *slog.Logger, port int) *App {
	return &App{}
}
