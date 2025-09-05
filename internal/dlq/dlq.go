package dlq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	skafka "github.com/segmentio/kafka-go"
	ikafka "github.com/sillkiw/wb-l0/internal/kafka" // твой тип Message
)

// Publisher — опционально, если хочешь иметь общий интерфейс для Close().
type Publisher interface {
	Send(ctx context.Context, m ikafka.Message, reason string, attempt int) error
	Close() error
}

type KafkaPublisher struct {
	w *skafka.Writer
}

func NewKafkaPublisher(brokers []string, topic string) (*KafkaPublisher, error) {
	if len(brokers) == 0 || brokers[0] == "" {
		return nil, errors.New("dlq: empty brokers")
	}
	if topic == "" {
		return nil, errors.New("dlq: empty topic")
	}
	return &KafkaPublisher{
		w: &skafka.Writer{
			Addr:         skafka.TCP(brokers...),
			Topic:        topic,
			RequiredAcks: skafka.RequireAll, // чтобы не потерять DLQ-сообщение
			Async:        false,
			Balancer:     &skafka.LeastBytes{},
		},
	}, nil
}

func (p *KafkaPublisher) Close() error { return p.w.Close() }

// Envelope — что положим в Value DLQ-сообщения.
type Envelope struct {
	OriginalTopic     string    `json:"original_topic"`
	OriginalPartition int       `json:"original_partition"`
	OriginalOffset    int64     `json:"original_offset"`
	Key               string    `json:"key"`     // исходный ключ сообщения (как строка)
	Reason            string    `json:"reason"`  // код причины (unmarshal_failed и т.п.)
	Attempt           int       `json:"attempt"` // счётчик попыток (если ведёшь)
	Payload           []byte    `json:"payload"` // исходный payload (base64 в JSON — это нормально)
	Timestamp         time.Time `json:"timestamp"`
}

func dlqKey(m ikafka.Message) []byte {
	// Ключ для дедупликации (удобно, если на DLQ включите compact,delete)
	return []byte(fmt.Sprintf("%s:%d:%d", m.Topic, m.Partition, m.Offset))
}

func (p *KafkaPublisher) Send(ctx context.Context, m ikafka.Message, reason string, attempt int) error {
	env := Envelope{
		OriginalTopic:     m.Topic,
		OriginalPartition: m.Partition,
		OriginalOffset:    m.Offset,
		Key:               string(m.Key),
		Reason:            reason,
		Attempt:           attempt,
		Payload:           m.Value,
		Timestamp:         time.Now().UTC(),
	}
	val, err := json.Marshal(env)
	if err != nil {
		return err
	}

	// Если ctx без дедлайна — дадим защитный таймаут, чтобы не зависнуть при shutdown.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 7*time.Second)
		defer cancel()
	}

	return p.w.WriteMessages(ctx, skafka.Message{
		Key:   dlqKey(m),
		Value: val,
		Headers: []skafka.Header{
			{Key: "x-dlq-reason", Value: []byte(reason)},
			{Key: "x-original-topic", Value: []byte(m.Topic)},
		},
	})
}

// Заглушка для dev/тестов
type NoopPublisher struct{}

func (NoopPublisher) Send(context.Context, ikafka.Message, string, int) error { return nil }
func (NoopPublisher) Close() error                                            { return nil }
