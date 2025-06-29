package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"goload/internal/generated/grpc/goload"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
}

func NewServer() Server {
	return &server{}
}

func (s *server) Start(ctx context.Context) error {
	mux := runtime.NewServeMux()
	if err := goload.RegisterGoLoadServiceHandlerFromEndpoint(
		ctx,
		mux,
		"/api",
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}); err != nil {
		return err
	}

	fmt.Println("http server is listening at port 8084")
	return http.ListenAndServe(":8084", mux)
}
