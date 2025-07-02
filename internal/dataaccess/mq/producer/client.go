package producer

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"goload/internal/configs"
	"goload/internal/utils"
)

var (
	produceMessageFailed = status.Error(codes.Internal, "failed to produce message")
)

type Client interface {
	Produce(ctx context.Context, topic string, payload []byte) error
}

type client struct {
	syncProducer sarama.SyncProducer
	logger       *zap.Logger
}

func NewClient(
	mqConfig configs.MQ,
	logger *zap.Logger,
) (Client, func(), error) {
	syncProducer, err := sarama.NewSyncProducer(mqConfig.Addresses, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create sarama sync producer: %w", err)
	}

	cleanup := func() {
		syncProducer.Close()
	}

	return &client{
		syncProducer: syncProducer,
		logger:       logger,
	}, cleanup, nil
}

// Produce implements Client.
func (c *client) Produce(ctx context.Context, topic string, payload []byte) error {
	logger := utils.LoggerWithContext(ctx, c.logger).
		With(zap.String("topic", topic)).
		With(zap.ByteString("payload", payload))

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(payload),
	}
	partition, offset, err := c.syncProducer.SendMessage(message)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to produce message")
		return produceMessageFailed
	}
	logger.
		With(zap.Int32("partition", partition)).
		With(zap.Int64("offset", offset)).
		Debug("message produced")

	return nil
}
