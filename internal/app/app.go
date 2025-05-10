package app

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"net"
	"sub-pub/internal/closer"
	"sub-pub/internal/config"
	"sub-pub/internal/logger"
	desc "sub-pub/pkg/subpub_v1"
	"time"
)

type App struct {
	serviceProvider *serviceProvider
	grpcServer      *grpc.Server
	logger          *zap.Logger
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Run() error {
	// Регистрируем функцию закрытия gRPC сервера
	closer.Add(func() error {
		a.logger.Info("Stopping gRPC server")
		if a.grpcServer != nil {
			// Создаем контекст с таймаутом для GracefulStop
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			// Запускаем GracefulStop в отдельной горутине
			done := make(chan struct{})
			go func() {
				a.grpcServer.GracefulStop()
				close(done)
			}()

			// Ждем либо завершения GracefulStop, либо таймаута
			select {
			case <-done:
				a.logger.Info("gRPC server stopped gracefully")
			case <-ctx.Done():
				a.logger.Warn("gRPC server graceful stop timeout, forcing stop")
				a.grpcServer.Stop()
			}
		}
		return nil
	})

	// Запускаем сервер
	a.logger.Info("Starting gRPC server",
		zap.String("address", a.serviceProvider.GRPCConfig().Address()),
	)

	list, err := net.Listen("tcp", a.serviceProvider.GRPCConfig().Address())
	if err != nil {
		return err
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		a.logger.Info("gRPC server is ready to serve")
		if err := a.grpcServer.Serve(list); err != nil {
			a.logger.Error("Failed to serve",
				zap.Error(err),
			)
		}
	}()

	// Ждем сигнала завершения
	a.logger.Info("Waiting for shutdown signal")
	<-closer.Wait()
	a.logger.Info("Application shutdown complete")
	return nil
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initLogger,
		a.initServiceProvider,
		a.initGRPCServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initConfig(_ context.Context) error {
	err := config.Load(".env")
	if err != nil {
		return err
	}

	return nil
}

func (a *App) initLogger(_ context.Context) error {
	logger, err := logger.NewLogger()
	if err != nil {
		return err
	}
	a.logger = logger
	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.serviceProvider = newServiceProvider()
	a.serviceProvider.SetLogger(a.logger)
	return nil
}

func (a *App) initGRPCServer(ctx context.Context) error {
	a.grpcServer = grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	reflection.Register(a.grpcServer)

	desc.RegisterPubSubV1Server(a.grpcServer, a.serviceProvider.SubPubImpl())

	return nil
}
