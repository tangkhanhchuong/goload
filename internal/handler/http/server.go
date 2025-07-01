package http

import (
	"context"
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
}

type server struct {
	httpConfig configs.HTTP
	grpcConfig configs.GRPC
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

	logger.With(zap.String("address", s.httpConfig.Address)).Info("http server is starting")
	return http.ListenAndServe(s.httpConfig.Address, mux)
}
