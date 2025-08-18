package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sillkiw/wb-l0/internal/app"
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := app.New(log, *addr)

	if err := a.Run(ctx); err != nil {
		log.Info("Server shutdown")
	}

}
