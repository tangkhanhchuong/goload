package grpc

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"goload/internal/configs"
	"goload/internal/generated/grpc/goload"
	"goload/internal/utils"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	handler    goload.GoLoadServiceServer
	grpcConfig configs.GRPC
	logger     *zap.Logger
}

func NewServer(
	handler goload.GoLoadServiceServer,
	grpcConfig configs.GRPC,
	logger *zap.Logger,
) Server {
	return &server{
		handler:    handler,
		grpcConfig: grpcConfig,
		logger:     logger,
	}
}

func (s *server) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	listener, err := net.Listen("tcp", s.grpcConfig.Address)
	if err != nil {
		return err
	}

	defer listener.Close()

	server := grpc.NewServer()
	goload.RegisterGoLoadServiceServer(server, s.handler)

	logger.With(zap.String("address", s.grpcConfig.Address)).Info("grpc server is starting")
	return server.Serve(listener)
}
