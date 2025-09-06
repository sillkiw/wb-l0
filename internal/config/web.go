package config

import (
	"log/slog"
	"time"
)

type WebConfig struct {
	AppEnv      Env
	Addr        string
	PostgresDSN string
	LogFormat   string
	LogLevel    string

	CacheSize int
	CacheTTL  time.Duration

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// LoadWeb читает настройки веб-приложения из окружения.
func LoadWeb() WebConfig {
	env := Env(get("APP_ENV", "local"))

	// Адрес HTTP сервера
	addr := get("HTTP_ADDR", ":4000")

	// Выбор DSN по окружению
	var dsn string
	switch env {
	case EnvDocker:
		dsn = get("DATABASE_URL", "")
	default: // local
		dsn = get("DATABASE_URL_HOST", "")
	}

	// Кэш
	cacheSize, ok1 := atoiDefault(get("CACHE_SIZE", "1000"), 1000)
	if !ok1 {
		slog.Warn("config: bad CACHE_SIZE, fallback to 1000")
	}
	cacheTTL, ok2 := durDefault(get("CACHE_TTL", "30s"), 30*time.Second)
	if !ok2 {
		slog.Warn("config: bad CACHE_TTL, fallback to 30s")
	}

	// Таймауты HTTP
	readTO, ok3 := durDefault(get("HTTP_READ_TIMEOUT", "10s"), 10*time.Second)
	if !ok3 {
		slog.Warn("config: bad HTTP_READ_TIMEOUT, fallback to 10s")
	}
	writeTO, ok4 := durDefault(get("HTTP_WRITE_TIMEOUT", "10s"), 10*time.Second)
	if !ok4 {
		slog.Warn("config: bad HTTP_WRITE_TIMEOUT, fallback to 10s")
	}
	idleTO, ok5 := durDefault(get("HTTP_IDLE_TIMEOUT", "60s"), 60*time.Second)
	if !ok5 {
		slog.Warn("config: bad HTTP_IDLE_TIMEOUT, fallback to 60s")
	}

	cfg := WebConfig{
		AppEnv:       env,
		Addr:         addr,
		PostgresDSN:  dsn,
		LogFormat:    get("LOG_FORMAT", "text"),
		LogLevel:     get("LOG_LEVEL", "INFO"),
		CacheSize:    cacheSize,
		CacheTTL:     cacheTTL,
		ReadTimeout:  readTO,
		WriteTimeout: writeTO,
		IdleTimeout:  idleTO,
	}

	// Лёгкие предупреждения
	if cfg.PostgresDSN == "" {
		slog.Warn("config: empty Postgres DSN", slog.String("APP_ENV", string(env)))
	}
	if cfg.Addr == "" {
		slog.Warn("config: empty HTTP_ADDR, fallback to :4000")
		cfg.Addr = ":4000"
	}

	return cfg
}
