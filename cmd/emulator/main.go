package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/sillkiw/wb-l0/internal/config"
	"github.com/sillkiw/wb-l0/internal/generator"
	"github.com/sillkiw/wb-l0/internal/kafka"
	"github.com/sillkiw/wb-l0/internal/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}
	cfg := config.LoadProducer()

	// log
	logg := logger.New(cfg.LogFormat, cfg.LogLevel)

	// продьюсер
	producer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer producer.Close()

	// Эмуляция некорректных сообщений
	cor := generator.NewCorruptor(cfg.BadRate, cfg.BadKinds)
	logg.Info("bad messages enabled",
		slog.Float64("rate", cfg.BadRate),
		slog.String("kinds", cfg.BadKinds),
	)

	logg.Info("Emulator started",
		slog.String("brokers", strings.Join(cfg.KafkaBrokers, ",")),
		slog.String("topic", cfg.KafkaTopic),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	for i := 0; cfg.Count == 0 || i < cfg.Count; i++ {
		select {
		case <-ctx.Done():
			logg.Info("Stopping producer")
			return
		default:
		}

		order := generator.NewFakeOrder(i)
		data, err := json.Marshal(order)
		if err != nil {
			logg.Warn("marshal failed", slog.Any("err", err))
			continue
		}

		// Возможно испортим сообщение
		data, reason := cor.Maybe(order, data)
		if reason != "" {
			logg.Warn("sending bad message",
				slog.String("order_uid", order.OrderUID),
				slog.String("reason", reason),
			)
		}

		sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = producer.SendWithContext(sendCtx, []byte(order.OrderUID), data)
		cancel()
		if err != nil {
			logg.Error("send failed", slog.Any("err", err))
		} else {
			logg.Debug("sent order", slog.String("order_uid", order.OrderUID))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(cfg.Interval):
		}
	}
}
