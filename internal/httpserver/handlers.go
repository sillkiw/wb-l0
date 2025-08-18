package httpserver

import (
	"log/slog"
	"net/http"
)

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.logger.Info("page not found",
			slog.String("path", r.URL.Path),
		)
		http.NotFound(w, r)
		return
	}

	s.logger.Info("request started",
		slog.String("ip", r.RemoteAddr),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("userAgent", r.UserAgent()),
	)
	s.serveUIFile(w, r, "index.html")
}
