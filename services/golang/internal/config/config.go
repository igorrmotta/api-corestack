package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL      string
	GRPCPort         string
	LogLevel         string
	RiverConcurrency int
}

func Load() *Config {
	concurrency, _ := strconv.Atoi(getEnv("RIVER_CONCURRENCY", "10"))

	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/api_corestack?sslmode=disable"),
		GRPCPort:         getEnv("GRPC_PORT", "8080"),
		LogLevel:         getEnv("LOG_LEVEL", "debug"),
		RiverConcurrency: concurrency,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
