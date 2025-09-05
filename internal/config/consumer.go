package config

import "log/slog"

type ConsumerConfig struct {
	AppEnv       Env
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
	DLQTopic     string
	PostgresDSN  string
	LogFormat    string
	LogLevel     string
}

func LoadConsumer() ConsumerConfig {
	env := Env(get("APP_ENV", "local"))
	b := selectBootstrap(env)

	cfg := ConsumerConfig{
		AppEnv:       env,
		KafkaBrokers: b.Brokers,
		KafkaTopic:   get("KAFKA_TOPIC", "orders"),
		KafkaGroupID: get("KAFKA_GROUP_ID", "orders-consumer"),
		DLQTopic:     get("KAFKA_DLQ_TOPIC", ""),
		PostgresDSN:  b.DSN,
		LogFormat:    get("LOG_FORMAT", "text"),
		LogLevel:     get("LOG_LEVEL", "INFO"),
	}
	if len(cfg.KafkaBrokers) == 0 {
		slog.Warn("config: empty Kafka bootstrap")
	}
	if cfg.PostgresDSN == "" {
		slog.Warn("config: empty Postgres DSN")
	}
	return cfg
}
