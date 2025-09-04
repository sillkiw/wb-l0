package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sillkiw/wb-l0/internal/domain"
	"github.com/sillkiw/wb-l0/internal/kafka"
	"github.com/sillkiw/wb-l0/internal/logger"
	"github.com/sillkiw/wb-l0/internal/storage"
)

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		ch := make(chan os.Signal, 2)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		<-ch
		os.Exit(1)
	}()

	// Логгер
	format := os.Getenv("LOG_FORMAT")
	level := os.Getenv("LOG_LEVEL")
	logg := logger.New(format, level)

	// kafka env
	broker := os.Getenv("KAFKA_BOOTSTRAP_EXTERNAL")
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")

	// БД
	dsn := os.Getenv("DATABASE_URL_HOST")
	store, err := storage.NewStorage(dsn)
	if err != nil {
		logg.Error("failed to connect to database",
			slog.Any("error", err))
		os.Exit(1)
	}
	defer store.Close()

	// Инициализация DLQ

	// Инициализация kafka-риддера
	r := kafka.NewReader([]string{broker}, topic, groupID)
	defer func() {
		if err := r.Close(); err != nil {
			logg.Error("failed to close Kafka reader", slog.Any("err", err))
		}
	}()

	logg.Info("Consumer started",
		slog.String("broker", broker),
		slog.String("topic", topic),
		slog.String("group_id", groupID),
	)

	for {
		// Чтение без коммита
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				logg.Info("shutdown: stop reading and exit loop")
				break
			}
			logg.Error("kafka read failed",
				slog.Any("err", err),
			)
			time.Sleep(time.Second)
			continue
		}
		start := time.Now()
		bLog := logg.With(
			slog.String("topic", msg.Topic),
			slog.Int("partition", msg.Partition),
			slog.Int64("offset", msg.Offset),
			slog.String("key", string(msg.Key)),
		)
		bLog.Debug("message received")

		// Десериализация
		var order domain.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			bLog.Warn("unmarshal failed; skipping message",
				slog.Any("err", err),
			)
			continue
		}
		// Запись в бд
		dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = store.SaveOrder(dbCtx, order)
		cancel()
		if err != nil {
			bLog.Error("failed to save order",
				slog.Any("err", err),
			)
			continue
		}
		bLog.Debug("order saved",
			slog.Duration("elapsed", time.Since(start)),
		)

		// Коммит офсета
		if err := commitWithTimeout(ctx, r, msg, bLog); err != nil {
			continue
		}
	}

	logg.Info("Consumer stopped",
		slog.String("broker", broker),
		slog.String("topic", topic),
		slog.String("group_id", groupID),
	)
}

// commitWithTimeout хелпер с таймаутом
func commitWithTimeout(
	parent context.Context,
	r *kafka.Consumer,
	msg kafka.Message,
	bLog *slog.Logger,
) error {
	commitCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if err := r.CommitMessages(commitCtx, msg); err != nil {
		bLog.Error("commit failed; message will be reprocessed",
			slog.Any("err", err),
		)
		return err
	}
	bLog.Debug("offset committed")
	return nil
}
