package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer struct{
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, group string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: 	brokers,
			Topic: 		topic,
			GroupID: 	group,
		}),
	}
}

func (c *Consumer) Read(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

