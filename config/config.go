package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort                 string
	AppEnvironment          string
	Timezone                string
	Zone                    string
	DBHost                  string
	DBPort                  string
	DBUser                  string
	DBPassword              string
	DBName                  string
	DBSSLMode               string
	DBMaxIdleConn           int
	DBMaxOpenConn           int
	DBMaxConnLifetimeMinute int
	RabbitMQURL             string
	RabbitMQQueue           string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Panic("Failed to load .env file. Make sure the .env file is available in the project root.")
	}

	cfg := &Config{
		AppPort:                 mustGetEnv("APP_PORT"),
		AppEnvironment:          mustGetEnv("APP_ENVIRONMENT"),
		Timezone:                mustGetEnv("TIMEZONE"),
		Zone:                    mustGetEnv("ZONE"),
		DBHost:                  mustGetEnv("DB_HOST"),
		DBPort:                  mustGetEnv("DB_PORT"),
		DBUser:                  mustGetEnv("DB_USER"),
		DBPassword:              mustGetEnv("DB_PASSWORD"),
		DBName:                  mustGetEnv("DB_NAME"),
		DBSSLMode:               mustGetEnv("DB_SSLMODE"),
		DBMaxIdleConn:           mustGetEnvInt("DB_MAX_IDLE_CONN"),
		DBMaxOpenConn:           mustGetEnvInt("DB_MAX_OPEN_CONN"),
		DBMaxConnLifetimeMinute: mustGetEnvInt("DB_CONN_MAX_LIFETIME"),
		RabbitMQURL:             mustGetEnv("RABBITMQ_URL"),
		RabbitMQQueue:           mustGetEnv("RABBITMQ_QUEUE"),
	}

	return cfg
}

func mustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		panic(fmt.Sprintf("Environment variable %s not found", key))
	}
	return value
}

func mustGetEnvInt(key string) int {
	valueStr := mustGetEnv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(fmt.Sprintf("Environment variable %s must be an integer", key))
	}
	return value
}
