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
	Partition int
	Offset    int64
	Topic     string
	Timestamp time.Time
}

func NewReader(brokers []string, topic string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	})
	return &Consumer{reader: reader}
}

func (c *Consumer) FetchMessage(ctx context.Context) (Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Key:       msg.Key,
		Value:     msg.Value,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Topic:     msg.Topic,
		Timestamp: msg.Time,
	}, nil
}

func (c *Consumer) CommitMessages(ctx context.Context, msgs ...Message) error {
	if len(msgs) == 0 {
		return nil
	}
	kmsgs := make([]kafka.Message, len(msgs))
	for i, m := range msgs {
		kmsgs[i] = kafka.Message{
			Topic:     m.Topic,
			Partition: m.Partition,
			Offset:    m.Offset,
		}
	}
	return c.reader.CommitMessages(ctx, kmsgs...)
}
func (c *Consumer) Close() error {
	return c.reader.Close()
}
