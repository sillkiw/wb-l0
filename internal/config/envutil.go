package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func atoiDefault(s string, def int) (int, bool) {
	if s == "" {
		return def, true
	}
	v, err := strconv.Atoi(s)
	return v, err == nil
}

func durDefault(s string, def time.Duration) (time.Duration, bool) {
	if s == "" {
		return def, true
	}
	d, err := time.ParseDuration(s)
	return d, err == nil
}
