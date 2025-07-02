package mq

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"goload/internal/dataaccess/mq/consumer"
	"goload/internal/dataaccess/mq/producer"
)

type MessageConsumer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type messageConsumer struct {
	downloadTaskCreatedHandler DownloadTaskCreated
	mqConsumer                 consumer.Consumer
	logger                     *zap.Logger
}

func NewMessageConsumer(
	downloadTaskCreatedHandler DownloadTaskCreated,
	mqConsumer consumer.Consumer,
	logger *zap.Logger,
) MessageConsumer {
	return &messageConsumer{
		downloadTaskCreatedHandler: downloadTaskCreatedHandler,
		mqConsumer:                 mqConsumer,
		logger:                     logger,
	}
}

func (r messageConsumer) Start(ctx context.Context) error {
	r.mqConsumer.RegisterHandler(
		producer.MessageQueueTopicDownloadTaskCreated,
		func(ctx context.Context, topic string, payload []byte) error {
			var event producer.DownloadTaskCreatedEvent
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}

			return r.downloadTaskCreatedHandler.Handle(ctx, event)
		},
	)

	return r.mqConsumer.Start(ctx)
}

func (r messageConsumer) Stop(ctx context.Context) error {
	return r.mqConsumer.Stop(ctx)
}
