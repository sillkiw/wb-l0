package httpserver

import (
	"log/slog"
	"net/http"
	"time"
)

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.log.Info("http",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Duration("dur", time.Since(start)),
			slog.String("ua", r.UserAgent()),
			slog.String("ip", r.RemoteAddr),
		)
	})
}
