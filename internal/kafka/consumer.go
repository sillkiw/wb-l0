package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

type Message struct {
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Partition int
	Offset    int64
	Topic     string
	Timestamp time.Time
}

func NewReader(brokers []string, topic string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &Consumer{reader: reader}
}

func (r *Consumer) ReadMessage(ctx context.Context) (m Message, err error) {
	msg, err := r.reader.ReadMessage(ctx)

	if err != nil {
		return Message{}, err
	}

	headers := make(map[string]string)
	for _, h := range msg.Headers {
		headers[h.Key] = string(h.Value)
	}

	return Message{
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   headers,
		Offset:    msg.Offset,
		Topic:     msg.Topic,
		Timestamp: msg.Time,
	}, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
