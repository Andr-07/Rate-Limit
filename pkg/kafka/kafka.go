package kafka

import (
	"context"
	"fmt"
	configs "rate-limiter/config"

	"github.com/IBM/sarama"
)

type KafkaClient struct {
	producer sarama.SyncProducer
}

type KafkaClientDeps struct {
	Config *configs.KafkaConfig
}

func NewKafkaClient(deps KafkaClientDeps) (*KafkaClient, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{deps.Config.Addr}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return &KafkaClient{producer: producer}, nil
}

func (k *KafkaClient) SendMessage(ctx context.Context, topic, message string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	_, _, err := k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %v", err)
	}

	return nil
}

func (k *KafkaClient) Close() {
	if err := k.producer.Close(); err != nil {
		fmt.Println("Failed to close Kafka producer:", err)
	}
}
