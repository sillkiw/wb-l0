package config

import (
	"log/slog"
	"strings"
	"time"
)

type ProducerConfig struct {
	AppEnv       Env
	KafkaBrokers []string
	KafkaTopic   string
	LogFormat    string
	LogLevel     string

	Count    int
	Interval time.Duration
	BadRate  float64
	BadKinds string
}

func LoadProducer() ProducerConfig {
	env := Env(get("APP_ENV", "local"))
	b := selectBootstrap(env)

	count, ok1 := atoiDefault(get("PRODUCER_COUNT", "0"), 0)
	if !ok1 {
		slog.Warn("config: bad PRODUCER_COUNT, use 0")
	}
	interval, ok2 := durDefault(get("PRODUCER_INTERVAL", "2s"), 2*time.Second)
	if !ok2 {
		slog.Warn("config: bad PRODUCER_INTERVAL, use 2s")
	}

	rate, ok3 := floatDefault(get("PRODUCER_BAD_RATE", "0"), 0)
	if !ok3 {
		slog.Warn("config: bad PRODUCER_BAD_RATE, use 0")
	}
	if rate < 0 {
		rate = 0
	} else if rate > 1 {
		rate = 1
	}
	kinds := strings.TrimSpace(get("PRODUCER_BAD_KINDS", ""))

	cfg := ProducerConfig{
		AppEnv:       env,
		KafkaBrokers: b.Brokers,
		KafkaTopic:   get("KAFKA_TOPIC", "orders"),
		LogFormat:    get("LOG_FORMAT", "text"),
		LogLevel:     get("LOG_LEVEL", "INFO"),
		Count:        count,
		Interval:     interval,
		BadRate:      rate,
		BadKinds:     kinds,
	}
	if len(cfg.KafkaBrokers) == 0 {
		slog.Warn("config: empty Kafka bootstrap")
	}
	return cfg
}
