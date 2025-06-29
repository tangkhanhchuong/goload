package grpc

import (
	"goload/internal/generated/grpc/goload"
)

type Handler struct {
	goload.UnimplementedGoLoadServiceServer
}

func NewHandler() goload.GoLoadServiceServer {
	return &Handler{}
}
