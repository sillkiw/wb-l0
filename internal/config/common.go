package config

type Env string

const (
	EnvLocal  Env = "local"
	EnvDocker Env = "docker"
)

type Bootstrap struct {
	Brokers []string
	DSN     string
}

func selectBootstrap(env Env) Bootstrap {
	switch env {
	case EnvDocker:
		return Bootstrap{
			Brokers: splitCSV(get("KAFKA_BOOTSTRAP_INTERNAL", "")),
			DSN:     get("DATABASE_URL", ""),
		}
	case EnvLocal:
		return Bootstrap{
			Brokers: splitCSV(get("KAFKA_BOOTSTRAP_EXTERNAL", "")),
			DSN:     get("DATABASE_URL", ""),
		}
	default:
		return Bootstrap{
			Brokers: splitCSV(get("KAFKA_BOOTSTRAP_EXTERNAL", "")),
			DSN:     get("DATABASE_URL_HOST", ""),
		}
	}
}
