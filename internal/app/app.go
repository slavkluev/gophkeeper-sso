package app

import (
	"time"

	"go.uber.org/zap"

	grpcapp "sso/internal/app/grpc"
	"sso/internal/service/auth"
	"sso/internal/storage/sqlite"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *zap.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
