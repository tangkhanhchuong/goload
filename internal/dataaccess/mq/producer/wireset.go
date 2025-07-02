package producer

import "github.com/google/wire"

var Wireset = wire.NewSet(
	NewClient,
	NewDownloadTaskCreatedProducer,
)
