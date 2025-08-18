package httpserver

import (
	"io/fs"
	"net/http"
	"path"
)

func (s *Server) serveUIFile(w http.ResponseWriter, r *http.Request, name string) {
	b, err := fs.ReadFile(s.ui, name) // name от корня s.ui, например "index.html"
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Простейшее определение Content-Type по расширению
	ext := path.Ext(name) // ".html", ".css", ".js"
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, _ = w.Write(b)
}
