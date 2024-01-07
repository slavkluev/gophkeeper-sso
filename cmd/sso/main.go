package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"sso/internal/app"
	"sso/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go func() {
		application.GRPCServer.MustRun()
	}()

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCServer.Stop()
	log.Info("Gracefully stopped")
}

func setupLogger(env string) *zap.Logger {
	var log *zap.Logger

	switch env {
	case envLocal:
		log, _ = zap.NewDevelopment()
	case envDev:
		log, _ = zap.NewDevelopment()
	case envProd:
		log, _ = zap.NewProduction()
	}

	defer log.Sync()

	return log
}
