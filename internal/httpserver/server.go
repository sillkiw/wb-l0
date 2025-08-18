package httpserver

import (
	"io/fs"
	"log/slog"
	"net/http"
)

type Server struct {
	mux    *http.ServeMux
	logger *slog.Logger
	ui     fs.FS // ФС со статикой
}

func New(logger *slog.Logger, ui fs.FS) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		logger: logger,
		ui:     ui,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.home)
	s.mux.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(
			http.FS(s.ui),
		),
		),
	)
}

func (s *Server) Handler() http.Handler { return s.mux }
