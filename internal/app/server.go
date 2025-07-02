package app

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"goload/internal/handler/grpc"
	"goload/internal/handler/http"
	"goload/internal/handler/mq"
)

type Server struct {
	grpcServer      grpc.Server
	httpServer      http.Server
	messageConsumer mq.MessageConsumer
	logger          *zap.Logger
}

func NewServer(
	grpcServer grpc.Server,
	httpServer http.Server,
	messageConsumer mq.MessageConsumer,
	logger *zap.Logger,
) *Server {
	return &Server{
		grpcServer:      grpcServer,
		httpServer:      httpServer,
		messageConsumer: messageConsumer,
		logger:          logger,
	}
}

func (s Server) Start() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		err := s.grpcServer.Start(ctx)
		s.logger.With(zap.Error(err)).Info("grpc server stopped")
	}()

	go func() {
		err := s.httpServer.Start(ctx)
		s.logger.With(zap.Error(err)).Info("http server stopped")
	}()

	go func() {
		consumerStartErr := s.messageConsumer.Start(ctx)
		s.logger.With(zap.Error(consumerStartErr)).Info("message queue consumer stopped")
	}()

	<-ctx.Done()
	s.logger.Info("shutting down...")

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := s.grpcServer.Stop(stopCtx); err != nil {
		s.logger.With(zap.Error(err)).Error("failed to stop grpc server")
	}
	if err := s.httpServer.Stop(stopCtx); err != nil {
		s.logger.With(zap.Error(err)).Error("failed to stop http server")
	}
	if err := s.messageConsumer.Stop(stopCtx); err != nil {
		s.logger.With(zap.Error(err)).Error("failed to stop message consumer")
	}
	s.logger.Info("shutdown complete")

	return nil
}
