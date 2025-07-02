package database

import (
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	InitializeDBAndMigrateUp,
	InitializeGoquDB,
	NewMigrator,
	NewAccountRepository,
	NewAccountPasswordRepository,
	NewPublicKeyRepository,
	NewDownloadRepository,
)
