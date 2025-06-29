package handler

import (
	"github.com/google/wire"

	"goload/internal/handler/grpc"
	"goload/internal/handler/http"
)

var WireSet = wire.NewSet(
	grpc.WireSet,
	http.WireSet,
)
