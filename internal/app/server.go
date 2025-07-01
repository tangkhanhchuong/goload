package app

import (
	"context"
	"syscall"

	"go.uber.org/zap"

	"goload/internal/handler/grpc"
	"goload/internal/handler/http"
	"goload/internal/utils"
)

type Server struct {
	grpcServer grpc.Server
	httpServer http.Server
	logger     *zap.Logger
}

func NewServer(
	grpcServer grpc.Server,
	httpServer http.Server,
	logger *zap.Logger,
) *Server {
	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
		logger:     logger,
	}
}

func (s Server) Start() error {
	go func() {
		err := s.grpcServer.Start(context.Background())
		s.logger.With(zap.Error(err)).Info("grpc server stopped")
	}()

	go func() {
		err := s.httpServer.Start(context.Background())
		s.logger.With(zap.Error(err)).Info("http server stopped")
	}()

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)
	return nil
}
