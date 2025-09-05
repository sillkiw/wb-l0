package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/sillkiw/wb-l0/internal/config"
	"github.com/sillkiw/wb-l0/internal/consumer"
	"github.com/sillkiw/wb-l0/internal/dlq"
	"github.com/sillkiw/wb-l0/internal/kafka"
	"github.com/sillkiw/wb-l0/internal/logger"
	"github.com/sillkiw/wb-l0/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}
	cfg := config.LoadConsumer()

	// log
	logg := logger.New(cfg.LogFormat, cfg.LogLevel)

	// бд
	store, err := storage.NewStorage(cfg.PostgresDSN)
	if err != nil {
		logg.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer store.Close()

	//  DLQ
	var dlqPub consumer.DLQ
	if cfg.DLQTopic != "" {
		pub, err := dlq.NewKafkaPublisher(cfg.KafkaBrokers, cfg.DLQTopic)
		if err != nil {
			logg.Error("DLQ init failed", slog.Any("err", err))
		} else {
			dlqPub = pub
			defer pub.Close()
			logg.Info("DLQ enabled", slog.String("topic", cfg.DLQTopic))
		}
	} else {
		logg.Warn("DLQ disabled (KAFKA_DLQ_TOPIC not set)")
	}

	// handler
	h := consumer.NewHandler(store, dlqPub, logg)

	//  kafka-риддер
	r := kafka.NewReader(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	defer func() {
		if err := r.Close(); err != nil {
			logg.Error("failed to close Kafka reader", slog.Any("err", err))
		}
	}()

	logg.Info("Consumer started",
		slog.String("broker", strings.Join(cfg.KafkaBrokers, ",")),
		slog.String("topic", cfg.KafkaTopic),
		slog.String("group_id", cfg.KafkaGroupID),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		ch := make(chan os.Signal, 2)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		<-ch
		os.Exit(1)
	}()

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				logg.Info("shutdown: stop reading and exit loop")
				break
			}
			logg.Error("kafka read failed", slog.Any("err", err))
			time.Sleep(time.Second)
			continue
		}

		if commit := h.Handle(ctx, msg); commit {
			_ = commitWithTimeout(context.Background(), r, msg, consumer.MsgLogger(logg, msg))
		}
	}

	logg.Info("Consumer stopped",
		slog.String("broker", strings.Join(cfg.KafkaBrokers, ",")),
		slog.String("topic", cfg.KafkaTopic),
		slog.String("group_id", cfg.KafkaGroupID),
	)
}

// commitWithTimeout хелпер с таймаутом
func commitWithTimeout(
	parent context.Context,
	r *kafka.Consumer,
	msg kafka.Message,
	log *slog.Logger,
) error {
	commitCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if err := r.CommitMessages(commitCtx, msg); err != nil {
		log.Error("commit failed; message will be reprocessed", slog.Any("err", err))
		return err
	}
	log.Debug("offset commited")
	return nil
}
