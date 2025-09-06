package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/sillkiw/wb-l0/internal/cache"
	"github.com/sillkiw/wb-l0/internal/domain"
)

type OrderStore interface {
	GetOrder(ctx context.Context, orderUID string) (domain.Order, error)
}

type Options struct {
	CacheSize    int
	CacheTTL     time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type Server struct {
	mux   *http.ServeMux
	log   *slog.Logger
	ui    http.FileSystem
	store OrderStore
	cache *cache.LRU[domain.Order]
}

func New(log *slog.Logger, store OrderStore, ui http.FileSystem, opts Options) *Server {
	if opts.CacheSize <= 0 {
		opts.CacheSize = 1000
	}
	if opts.CacheTTL <= 0 {
		opts.CacheTTL = 30 * time.Second
	}

	s := &Server{
		mux:   http.NewServeMux(),
		log:   log,
		ui:    ui,
		store: store,
		cache: cache.NewLRU[domain.Order](opts.CacheSize, opts.CacheTTL),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.withLogging(s.mux) }

func (s *Server) routes() {
	// UI: статика и главная
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(s.ui)))
	s.mux.HandleFunc("/", s.handleIndex)

	// API
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /order/", s.handleGetOrder) // /api/orders/{order_uid}
}
