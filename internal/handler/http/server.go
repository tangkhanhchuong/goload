package http

import (
	"context"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"goload/internal/configs"
	"goload/internal/generated/grpc/goload"
	"goload/internal/utils"
)

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type server struct {
	httpConfig configs.HTTP
	grpcConfig configs.GRPC
	httpServer *http.Server
	logger     *zap.Logger
}

func NewServer(
	httpConfig configs.HTTP,
	grpcConfig configs.GRPC,
	logger *zap.Logger,
) Server {
	return &server{
		httpConfig: httpConfig,
		grpcConfig: grpcConfig,
		logger:     logger,
	}
}

func (s *server) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	mux := runtime.NewServeMux()
	if err := goload.RegisterGoLoadServiceHandlerFromEndpoint(
		ctx,
		mux,
		s.grpcConfig.Address,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", s.httpConfig.Address)
	if err != nil {
		return err
	}
	s.httpServer = &http.Server{
		Handler: mux,
	}

	logger.With(zap.String("address", s.httpConfig.Address)).Info("http server is starting")
	return s.httpServer.Serve(listener)
}

func (s *server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
