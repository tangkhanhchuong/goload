package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"goload/internal/generated/grpc/goload"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	handler goload.GoLoadServiceServer
}

func NewServer(
	handler goload.GoLoadServiceServer,
) Server {
	return &server{
		handler: handler,
	}
}

func (s *server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", "localhost:8083")
	if err != nil {
		return err
	}

	defer listener.Close()

	server := grpc.NewServer()
	goload.RegisterGoLoadServiceServer(server, s.handler)

	fmt.Println("grpc server is listening at port 8083")
	return server.Serve(listener)
}
