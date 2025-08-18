package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/sillkiw/wb-l0/internal/httpserver"
)

type App struct {
	log *slog.Logger
	srv *http.Server
}

func New(log *slog.Logger, httpAddr string) *App {
	ui := httpserver.EmbeddedUI()
	hs := httpserver.New(log, ui)
	errlog := slog.NewLogLogger(log.Handler(), slog.LevelError)
	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      hs.Handler(),
		ErrorLog:     errlog,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return &App{log: log, srv: srv}
}

func (a *App) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		sh, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = a.srv.Shutdown(sh)
	}()
	a.log.Info("http listen", slog.String("addr", a.srv.Addr))
	if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
