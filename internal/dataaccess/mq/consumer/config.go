package consumer

import (
	"github.com/IBM/sarama"

	"goload/internal/configs"
)

func newSaramaConfig(mqConfig configs.MQ) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true
	return saramaConfig
}
