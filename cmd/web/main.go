// cmd/web/main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/sillkiw/wb-l0/internal/app"
	"github.com/sillkiw/wb-l0/internal/config"
	"github.com/sillkiw/wb-l0/internal/httpserver"
	"github.com/sillkiw/wb-l0/internal/logger"
	"github.com/sillkiw/wb-l0/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}
	// Конфиг и логгер
	web := config.LoadWeb()
	log := logger.New(web.LogFormat, web.LogLevel)

	// Подключение к БД
	store, err := storage.NewStorage(web.PostgresDSN)
	if err != nil {
		log.Error("db connect failed", slog.Any("err", err))
		os.Exit(1)
	}
	defer store.Close()

	// UI
	var ui http.FileSystem
	if web.AppEnv == config.EnvLocal {
		ui = http.Dir("static")
	} else {
		ui = httpserver.EmbeddedUI()
	}

	// HTTP сервер с роутами и кэшем
	hs := httpserver.New(
		log,
		store, // реализует httpserver.OrderStore
		ui,
		httpserver.Options{
			CacheSize: web.CacheSize,
			CacheTTL:  web.CacheTTL,
		},
	)

	a := app.New(
		log,
		app.Config{
			Addr:         web.Addr,
			ReadTimeout:  web.ReadTimeout,
			WriteTimeout: web.WriteTimeout,
			IdleTimeout:  web.IdleTimeout,
		},
		hs,
	)

	// Завершаем по Ctrl+C / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := a.Run(ctx); err != nil {
		log.Error("server error", slog.Any("err", err))
	}
}
