package config

import "os"

type Config struct {
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	HMSAPIKey       string
	HMSAPISecret    string
	Environment     string
	FrontendURL     string
}

func Load() *Config {
	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost/psycho_platform?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		HMSAPIKey:    getEnv("HMS_API_KEY", ""),
		HMSAPISecret: getEnv("HMS_API_SECRET", ""),
		Environment:  getEnv("ENVIRONMENT", "development"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
