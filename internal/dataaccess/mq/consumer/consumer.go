package consumer

import (
	"context"
	"fmt"
	"goload/internal/configs"
	"goload/internal/utils"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type Consumer interface {
	RegisterHandler(topic string, handleFunc HandlerFunc)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type consumer struct {
	saramaConsumer        sarama.ConsumerGroup
	topicToHandlerFuncMap map[string]HandlerFunc
	logger                *zap.Logger
	cancelFunc            context.CancelFunc
}

func NewConsumer(
	mqConfig configs.MQ,
	logger *zap.Logger,
) (Consumer, error) {
	saramaConsumer, err := sarama.NewConsumerGroup(mqConfig.Addresses, mqConfig.ClientID, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}
	logger.Info("consumer connected")

	return &consumer{
		saramaConsumer:        saramaConsumer,
		logger:                logger,
		topicToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}

// RegisterHandler implements Consumer.
func (c *consumer) RegisterHandler(topic string, handleFunc HandlerFunc) {
	c.topicToHandlerFuncMap[topic] = handleFunc
}

// Start implements Consumer.
func (c *consumer) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, c.logger)

	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	for topic, handlerFunc := range c.topicToHandlerFuncMap {
		go func(topic string, handlerFunc HandlerFunc) {
			err := c.saramaConsumer.Consume(
				context.Background(),
				[]string{topic},
				newConsumerHandler(handlerFunc),
			)
			if err != nil {
				logger.
					With(zap.String("topic", topic)).
					With(zap.Error(err)).
					Error("failed to consume message from queue")
			}
		}(topic, handlerFunc)
	}
	<-ctx.Done()

	return nil
}

func (c *consumer) Stop(ctx context.Context) error {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	// Gracefully close the sarama consumer
	return c.saramaConsumer.Close()
}
