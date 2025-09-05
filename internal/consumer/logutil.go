package consumer

import (
	"log/slog"

	"github.com/sillkiw/wb-l0/internal/kafka"
)

func MsgLogger(base *slog.Logger, msg kafka.Message) *slog.Logger {
	return base.With(
		slog.String("topic", msg.Topic),
		slog.Int("partition", msg.Partition),
		slog.Int64("offset", msg.Offset),
		slog.String("key", string(msg.Key)),
	)
}
