package dataaccess

import (
	"github.com/google/wire"

	"goload/internal/dataaccess/database"
	"goload/internal/dataaccess/mq"
)

var WireSet = wire.NewSet(
	database.WireSet,
	mq.WireSet,
)
