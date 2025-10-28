package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort       string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	RabbitMQURL   string
	RabbitMQQueue string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:       getEnv("APP_PORT", "8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "postgres"),
		DBName:        getEnv("DB_NAME", "jobdb"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		RabbitMQURL:   getEnv("RABBITMQ_URL", ""),
		RabbitMQQueue: getEnv("RABBITMQ_QUEUE", "job_queue"),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
