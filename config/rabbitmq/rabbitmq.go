package rabbitmq

import (
	"background-job-service/config"
	"background-job-service/config/logger"
	"crypto/tls"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Credential struct {
	User     string `json:"user" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type DSN struct {
	Host string `json:"host" validate:"required"`
	Port string `json:"port" validate:"required"`
}

type HealthCheckConfig struct {
	HealthCheckInterval       uint8
	ReconnectMaxRetries       uint8
	ReconnectInterval         uint8
	CircuitBreakerMaxFailures uint8
	CircuitBreakerTimeout     uint8
}

type Config struct {
	Vhost           string        `json:"vhost" `
	Heartbeat       time.Duration `json:"heartbeat"` // in Seconds
	TLSClientConfig *tls.Config   `json:"tls_config"`

	HealthCheckConfig
}

type Connection struct {
	DSN        DSN        `json:"dsn"`
	Credential Credential `json:"credential"`
	Config     Config     `json:"config"`
}

type RabbitMQ struct {
	Connection Connection `json:"connection" validate:"required"`
	Logger     *logger.Logger
}

func NewRabbitMQConfig(cfg *config.Config, l *logger.Logger, tlsConfig *tls.Config) *RabbitMQ {
	return &RabbitMQ{
		Connection: Connection{
			DSN: DSN{
				Host: cfg.RabbitMQ.Host,
				Port: cfg.RabbitMQ.Port,
			},
			Credential: Credential{
				User:     cfg.RabbitMQ.User,
				Password: cfg.RabbitMQ.Password,
			},
			Config: Config{
				Vhost:           cfg.RabbitMQ.VHost,
				Heartbeat:       time.Duration(cfg.RabbitMQ.HearthbeatInterval) * time.Second,
				TLSClientConfig: tlsConfig,
				HealthCheckConfig: HealthCheckConfig{
					HealthCheckInterval:       cfg.RabbitMQ.HealthCheckInterval,
					ReconnectMaxRetries:       cfg.RabbitMQ.ReconnectMaxRetries,
					ReconnectInterval:         cfg.RabbitMQ.ReconnectInterval,
					CircuitBreakerMaxFailures: cfg.RabbitMQ.CircuitBreakerMaxFailures,
					CircuitBreakerTimeout:     cfg.RabbitMQ.HealthCheckInterval,
				},
			},
		},

		Logger: l,
	}
}

func (r *RabbitMQ) Connect() (*amqp.Connection, error) {
	dsn := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/%s",
		r.Connection.Credential.User,
		r.Connection.Credential.Password,
		r.Connection.DSN.Host,
		r.Connection.DSN.Port,
		r.Connection.Config.Vhost,
	)

	config := amqp.Config{
		Heartbeat:       r.Connection.Config.Heartbeat * time.Second,
		TLSClientConfig: r.Connection.Config.TLSClientConfig,
	}

	return amqp.DialConfig(dsn, config)
}
