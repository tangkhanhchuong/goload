package producer

import (
	"github.com/IBM/sarama"

	"goload/internal/configs"
)

func newSaramaConfig(mqConfig configs.MQ) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Retry.Max = 1
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true
	return saramaConfig
}
