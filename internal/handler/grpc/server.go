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
	Stop(ctx context.Context) error
}

type server struct {
	handler    goload.GoLoadServiceServer
	grpcConfig configs.GRPC
	grpcServer *grpc.Server
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

	server := grpc.NewServer()
	s.grpcServer = server
	goload.RegisterGoLoadServiceServer(server, s.handler)

	logger.With(zap.String("address", s.grpcConfig.Address)).Info("grpc server is starting")
	return server.Serve(listener)
}

// Stop implements Server.
func (s *server) Stop(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		if s.grpcServer != nil {
			s.grpcServer.GracefulStop()
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
