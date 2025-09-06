// internal/app/app.go
package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/sillkiw/wb-l0/internal/httpserver"
)

type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type App struct {
	log *slog.Logger
	srv *http.Server
}

// New собирает http.Server, используя Handler из httpserver.Server.
func New(log *slog.Logger, cfg Config, hs *httpserver.Server) *App {
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 60 * time.Second
	}

	errlog := slog.NewLogLogger(log.Handler(), slog.LevelError)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      hs.Handler(),
		ErrorLog:     errlog,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &App{
		log: log,
		srv: srv,
	}
}

// Run запускает HTTP-сервер и останавливает его по ctx.Done() с таймаутом.
func (a *App) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.srv.Shutdown(shCtx); err != nil {
			a.log.Error("http shutdown", slog.Any("err", err))
		}
	}()

	a.log.Info("http listen", slog.String("addr", a.srv.Addr))

	if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
