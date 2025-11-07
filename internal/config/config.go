package config

import (
	"os"
	"strconv"
)

type Config struct {
	// LogIngestor config
	LogIngestorAddr string

	// MetricsIngestor config
	MetricsIngestorAddr string

	// Database config
	PostgresURL  string
	TimescaleURL string

	// NATS config
	NATSURL string

	// API server config
	APIServerAddr string
}

func Load() *Config {
	return &Config{
		LogIngestorAddr:     getEnv("LOG_INGESTOR_ADDR", ":8080"),
		MetricsIngestorAddr: getEnv("METRICS_INGESTOR_ADDR", ":8081"),
		PostgresURL:         getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/logs?sslmode=disable"),
		TimescaleURL:        getEnv("TIMESCALE_URL", "postgres://user:password@localhost:5432/metrics?sslmode=disable"),
		NATSURL:             getEnv("NATS_URL", "nats://localhost:4222"),
		APIServerAddr:       getEnv("API_SERVER_ADDR", ":8082"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
