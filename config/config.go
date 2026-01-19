package config

import (
	"background-job-service/pkg/constant"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName                 string
	AppPort                 string
	AppEnvironment          string
	Timezone                string
	TimeoutGracefulShutdown uint8

	DB
	RabbitMQ
	Logstash
	Worker
}

type DB struct {
	DBHost                  string
	DBPort                  string
	DBUser                  string
	DBPassword              string
	DBName                  string
	DBSSLMode               string
	DBMaxIdleConn           int
	DBMaxOpenConn           int
	DBMaxConnLifetimeMinute int
}

type RabbitMQ struct {
	Host                      string
	Port                      string
	User                      string
	Password                  string
	VHost                     string
	HearthbeatInterval        uint8
	HealthCheckInterval       uint8
	ReconnectMaxRetries       uint8
	ReconnectInterval         uint8
	CircuitBreakerMaxFailures uint8
	CircuitBreakerTimeout     uint8
}

type Logstash struct {
	Host              string
	Port              string
	Timeout           uint8
	TickerHealthCheck uint8
}

type Worker struct {
	MultiPostingWorker
}

type MultiPostingWorker struct {
	Enabled      bool
	RetryEnabled bool
	MaxRetry     uint8
	Concurrency  uint8
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Panic("Failed to load .env file. Make sure the .env file is available in the project root.")
	}

	cfg := &Config{
		AppName:                 mustGetEnv(constant.APP_NAME),
		AppPort:                 mustGetEnv(constant.APP_PORT),
		AppEnvironment:          mustGetEnv(constant.APP_ENVIRONMENT),
		Timezone:                mustGetEnv(constant.TIMEZONE),
		TimeoutGracefulShutdown: mustGetEnvUint8(constant.TIMEOUT_GRACEFUL_SHUTDOWN),
		DB: DB{
			DBHost:                  mustGetEnv(constant.DB_HOST),
			DBPort:                  mustGetEnv(constant.DB_PORT),
			DBUser:                  mustGetEnv(constant.DB_USER),
			DBPassword:              mustGetEnv(constant.DB_PASSWORD),
			DBName:                  mustGetEnv(constant.DB_NAME),
			DBSSLMode:               mustGetEnv(constant.DB_SSLMODE),
			DBMaxIdleConn:           mustGetEnvInt(constant.DB_MAX_IDLE_CONN),
			DBMaxOpenConn:           mustGetEnvInt(constant.DB_MAX_OPEN_CONN),
			DBMaxConnLifetimeMinute: mustGetEnvInt(constant.DB_CONN_MAX_LIFETIME),
		},
		RabbitMQ: RabbitMQ{
			Host:                      mustGetEnv(constant.RABBITMQ_HOST),
			Port:                      mustGetEnv(constant.RABBITMQ_PORT),
			User:                      mustGetEnv(constant.RABBITMQ_USER),
			Password:                  mustGetEnv(constant.RABBITMQ_PASSWORD),
			VHost:                     mustGetEnv(constant.RABBITMQ_VHOST),
			HearthbeatInterval:        mustGetEnvUint8(constant.RABBITMQ_HEARTBEAT_INTERVAL),
			HealthCheckInterval:       mustGetEnvUint8(constant.RABBITMQ_HEALTH_CHECK_INTERVAL),
			ReconnectMaxRetries:       mustGetEnvUint8(constant.RABBITMQ_RECONNECT_MAX_RETRIES),
			ReconnectInterval:         mustGetEnvUint8(constant.RABBITMQ_RECONNECT_INTERVAL),
			CircuitBreakerMaxFailures: mustGetEnvUint8(constant.RABBITMQ_CIRCUIT_BREAKER_MAX_FAILURES),
			CircuitBreakerTimeout:     mustGetEnvUint8(constant.RABBITMQ_CIRCUIT_BREAKER_TIMEOUT),
		},
		Logstash: Logstash{
			Host:              mustGetEnv(constant.LOGSTASH_HOST),
			Port:              mustGetEnv(constant.LOGSTASH_PORT),
			Timeout:           mustGetEnvUint8(constant.LOGSTASH_TIMEOUT),
			TickerHealthCheck: mustGetEnvUint8(constant.LOGSTASH_TICKER_HEALTHCHECK),
		},
		Worker: Worker{
			MultiPostingWorker{
				Enabled:      mustGetEnvBool(constant.WORKER_MULTI_POSTING_ENABLED),
				RetryEnabled: mustGetEnvBool(constant.WORKER_MULTI_POSTING_RETRY_ENABLED),
				MaxRetry:     mustGetEnvUint8(constant.WORKER_MULTI_POSTING_MAX_RETRY),
				Concurrency:  mustGetEnvUint8(constant.WORKER_MULTI_POSTING_CONCURRENCY),
			},
		},
	}

	return cfg
}

func mustGetEnvBool(key string) bool {
	valueStr := mustGetEnv(key)

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		panic(fmt.Sprintf("Environment variable %s must be a boolean (true/false)", key))
	}

	return value
}

func mustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		panic(fmt.Sprintf("Environment variable %s not found", key))
	}
	return value
}

func mustGetEnvUint8(key string) uint8 {
	valueStr := mustGetEnv(key)

	value64, err := strconv.ParseUint(valueStr, 10, 8)
	if err != nil {
		panic(fmt.Sprintf("Environment variable %s must be a uint8", key))
	}

	return uint8(value64)
}

func mustGetEnvInt(key string) int {
	valueStr := mustGetEnv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(fmt.Sprintf("Environment variable %s must be an integer", key))
	}
	return value
}
