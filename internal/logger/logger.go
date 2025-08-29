package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func New(format string, level string) *slog.Logger {
	var slogLevel slog.Level
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "INFO":
		slogLevel = slog.LevelInfo
	case "WARN":
		slogLevel = slog.LevelWarn
	case "ERROR":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Value = slog.TimeValue(time.Now().UTC())
			case slog.SourceKey:
				if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
					file := shortPath(src.File, 2)
					return slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", file, src.Line))
				}
			}
			return a
		},
	}

	var l slog.Handler
	if strings.ToLower(format) == "json" {
		l = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		l = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(l)
}

func shortPath(p string, keep int) string {
	parts := strings.Split(p, string(os.PathSeparator))
	if keep <= 0 || len(parts) <= keep {
		return filepath.ToSlash(filepath.Clean(p))
	}
	return strings.Join(parts[len(parts)-keep:], "/")
}
