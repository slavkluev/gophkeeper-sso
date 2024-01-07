package grpc

import (
	"fmt"
	"net"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authgrpc "sso/internal/grpc/auth"
)

type App struct {
	log        *zap.Logger
	GRPCServer *grpc.Server
	port       int
}

func New(
	log *zap.Logger,
	authService authgrpc.Auth,
	port int,
) *App {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", zap.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		grpczap.UnaryServerInterceptor(log),
	))

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		GRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", zap.String("addr", l.Addr().String()))

	if err := a.GRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(zap.String("op", op)).
		Info("stopping gRPC server", zap.Int("port", a.port))

	a.GRPCServer.GracefulStop()
}
