package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sillkiw/wb-l0/internal/kafka"
	"github.com/sillkiw/wb-l0/internal/logger"
)

func main() {
	// Загрузка файла с переменными окружения
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	// Инициализация логгера
	format := os.Getenv("LOG_FORMAT")
	level := os.Getenv("LOG_LEVEL")
	logger := logger.New(format, level)

	// Получение переменных с .env файла
	broker := os.Getenv("KAFKA_EXTERNAL")
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")

	// Инициализация kafka-риддера
	r := kafka.NewReader([]string{broker}, topic, groupID)
	defer func() {
		if err := r.Close(); err != nil {
			logger.Error("failed to close Kafka reader", slog.Any("err", err))
		}
	}()

	logger.Info("Consumer started",
		slog.String("broker", broker),
		slog.String("topic", topic),
		slog.String("GROUP_ID", groupID),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			logger.Error("read failed",
				slog.Any("err", err),
			)
			time.Sleep(time.Second)
			continue
		}
		logger.Info("get message from kafka",
			slog.String("topic", msg.Topic),
			slog.Int("partition", msg.Partition),
			slog.Int64("offset", msg.Offset),
			slog.Time("record_time", msg.Timestamp.UTC()),
		)
		logger.Debug("message from kafka",
			slog.String("key", string(msg.Key)),
			slog.String("value", string(msg.Value)),
		)

	}
}
