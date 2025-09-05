package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/sillkiw/wb-l0/internal/domain"
	"github.com/sillkiw/wb-l0/internal/kafka"
	"github.com/sillkiw/wb-l0/internal/validation"
)

type Store interface {
	SaveOrder(ctx context.Context, order domain.Order) error
}

type DLQ interface {
	Send(ctx context.Context, m kafka.Message, reason string, attempt int) error
}

type Handler struct {
	store Store
	dlq   DLQ
	log   *slog.Logger
}

func NewHandler(store Store, dlq DLQ, log *slog.Logger) *Handler {
	return &Handler{store: store, dlq: dlq, log: log}
}

func (h *Handler) Handle(ctx context.Context, msg kafka.Message) bool {
	log := MsgLogger(h.log, msg)

	var order domain.Order
	log.Debug("message", slog.String("valuer", string(msg.Value)))

	// строгий парсинг
	if err := validation.DecodeStrict(msg.Value, &order); err != nil {
		log.Warn("json decode failed", slog.Any("err", err))
		if h.dlq != nil {
			if err2 := h.dlq.Send(context.Background(), msg, "unmarshal_failed", 1); err2 != nil {
				log.Error("dlq send failed", slog.Any("err", err2))
				return false // DLQ временно недоступен - ретраим
			}
			log.Debug("sent to DLQ", slog.String("reason", "unmarshal_failed"))
			return true // коммитим, чтобы не зациклиться
		}
		return false // без DLQ - ретраим
	}

	// валидация
	if verr := validation.ValidateOrder(order); verr != nil {
		log.Warn("validation failed",
			slog.Int("errors", len(verr.Fields)),
			slog.String("summary", verr.Error()),
		)
		if h.dlq != nil {
			if err2 := h.dlq.Send(context.Background(), msg, "validation_failed", 1); err2 != nil {
				log.Error("dlq send failed", slog.Any("err", err2))
				return false
			}
			log.Debug("sent to DLQ", slog.String("reason", "validation_failed"))
			return true
		}
		return false
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.store.SaveOrder(dbCtx, order); err != nil {
		log.Error("save failed", slog.Any("err", err))
		return false
	}
	log.Debug("order saved")

	return true
}
