package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sillkiw/wb-l0/internal/storage"
)

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/order/")
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "bad order id", http.StatusBadRequest)
		return
	}

	// cache
	if o, ok := s.cache.Get(id); ok {
		w.Header().Set("X-Source", "cache")
		w.Header().Set("Content-Type", "application/json")
		s.respondJSON(w, http.StatusOK, o)
		return
	}

	// db
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	o, err := s.store.GetOrder(ctx, id)
	if err != nil {
		w.Header().Set("X-Source", "miss")
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		s.log.Error("get order failed",
			slog.String("order_uid", id),
			slog.Any("err", err),
		)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Source", "db")
	w.Header().Set("Content-Type", "application/json")
	s.cache.Set(id, o)
	s.respondJSON(w, http.StatusOK, o)
}

func (s *Server) respondJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
