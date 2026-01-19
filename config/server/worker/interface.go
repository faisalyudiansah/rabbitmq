package worker

import (
	"background-job-service/config/logger"
	"background-job-service/config/rabbitmq"
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Worker interface that all workers must implement
type Worker interface {
	// GetQueueName returns the queue name this worker listens to
	GetQueueName() string

	// IsEnabled checks if this worker is enabled via config
	IsEnabled() bool

	// IsRetryEnabled checks if retry is enabled for this worker
	IsRetryEnabled() bool

	// GetMaxRetry returns maximum retry attempts
	GetMaxRetry() int

	// GetConcurrency returns number of concurrent workers
	GetConcurrency() int

	// Process handles the message
	Process(ctx context.Context, delivery amqp.Delivery) error

	// OnError is called when processing fails
	OnError(ctx context.Context, delivery amqp.Delivery, err error)

	// OnSuccess is called when processing succeeds
	OnSuccess(ctx context.Context, delivery amqp.Delivery)
}

// BaseWorker provides common functionality
type BaseWorker struct {
	QueueName    string
	Logger       *logger.Logger
	QueueWrapper *rabbitmq.RabbitMQWrapper
	WorkerLogger *WorkerLogger
	Enabled      bool
	RetryEnabled bool
	MaxRetry     int
	Concurrency  int
}

func (b *BaseWorker) GetQueueName() string {
	return b.QueueName
}

func (b *BaseWorker) IsEnabled() bool {
	return b.Enabled
}

func (b *BaseWorker) IsRetryEnabled() bool {
	return b.RetryEnabled
}

func (b *BaseWorker) GetMaxRetry() int {
	return b.MaxRetry
}

func (b *BaseWorker) GetConcurrency() int {
	if b.Concurrency <= 0 {
		return 1 // default concurrency
	}
	return b.Concurrency
}

func (b *BaseWorker) OnError(ctx context.Context, delivery amqp.Delivery, err error) {
	b.Logger.Logger.WithError(err).Errorf(
		"[%s] Message processing failed: %s",
		b.QueueName,
		string(delivery.Body),
	)
}

func (b *BaseWorker) OnSuccess(ctx context.Context, delivery amqp.Delivery) {
	b.Logger.Logger.Infof(
		"[%s] Message processed successfully",
		b.QueueName,
	)
}
