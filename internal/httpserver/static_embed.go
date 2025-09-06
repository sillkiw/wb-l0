package httpserver

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var embeddedStatic embed.FS

// EmbeddedUI возвращает http.FileSystem поверх встроенной статики.
func EmbeddedUI() http.FileSystem {
	sub, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
