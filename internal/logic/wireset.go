package logic

import (
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	NewAccountService,
	NewHashService,
	NewTokenService,
	NewDownloadTaskService,
)
