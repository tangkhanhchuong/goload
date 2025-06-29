package dataaccess

import (
	"github.com/google/wire"

	"goload/internal/dataaccess/database"
)

var WireSet = wire.NewSet(
	database.WireSet,
)
