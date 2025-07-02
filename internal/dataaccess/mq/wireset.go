package mq

import (
	"github.com/google/wire"

	"goload/internal/dataaccess/mq/consumer"
	"goload/internal/dataaccess/mq/producer"
)

var WireSet = wire.NewSet(
	consumer.WireSet,
	producer.Wireset,
)
