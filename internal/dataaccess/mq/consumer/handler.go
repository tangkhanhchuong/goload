package consumer

import (
	"context"

	"github.com/IBM/sarama"
)

type HandlerFunc func(ctx context.Context, topic string, payload []byte) error

type consumerHandler struct {
	handlerFunc HandlerFunc
}

func newConsumerHandler(
	handlerFunc HandlerFunc,
) *consumerHandler {
	return &consumerHandler{
		handlerFunc: handlerFunc,
	}
}

// Cleanup implements sarama.ConsumerGroupHandler.
func (c *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// Setup implements sarama.ConsumerGroupHandler.
func (c *consumerHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim implements sarama.ConsumerGroupHandler.
func (c *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				session.Commit()
				return nil
			}

			if err := c.handlerFunc(session.Context(), message.Topic, message.Value); err != nil {
				return err
			}
		case <-session.Context().Done():
			return nil
		}
	}
}
