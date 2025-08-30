package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sillkiw/wb-l0/internal/domain"
	"github.com/sillkiw/wb-l0/internal/kafka"
	"github.com/sillkiw/wb-l0/internal/logger"
	"github.com/sillkiw/wb-l0/internal/storage"
)

func main() {
	// Загрузка файла с переменными окружения
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация логгера
	format := os.Getenv("LOG_FORMAT")
	level := os.Getenv("LOG_LEVEL")
	logger := logger.New(format, level)

	// Получение переменных для kafka с .env файла
	broker := os.Getenv("KAFKA_EXTERNAL")
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")

	// БД
	connStr := os.Getenv("POSTGRES_DSN_EXTERNAL")
	store, err := storage.NewStorage(connStr)
	if err != nil {
		logger.Error("failed to connect to database",
			slog.Any("error", err))
	}

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
		order := domain.Order{}
		err = json.Unmarshal(msg.Value, &order)
		if err != nil {
			logger.Warn("failed to marshal message",
				slog.Any("err", err))
		}
		if err := store.SaveOrder(ctx, order); err != nil {
			logger.Error("failed to save order to db",
				slog.Any("err", err))
		}
	}
}
