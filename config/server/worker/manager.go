package worker

import (
	gommonlog "background-job-service/config/gommon"
	"background-job-service/config/rabbitmq"
	"background-job-service/pkg/dto"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type WorkerManager struct {
	workers      []Worker
	logger       *gommonlog.Logger
	queueWrapper *rabbitmq.RabbitMQWrapper
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	consumerWg   sync.WaitGroup
}

func NewWorkerManager(
	logger *gommonlog.Logger,
	queueWrapper *rabbitmq.RabbitMQWrapper,
) *WorkerManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerManager{
		workers:      make([]Worker, 0),
		logger:       logger,
		queueWrapper: queueWrapper,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// RegisterWorker adds a worker to the manager
func (m *WorkerManager) RegisterWorker(worker Worker) error {
	if !worker.IsEnabled() {
		m.logger.Infof(
			"[-] Worker [%s] is disabled, skipping registration",
			worker.GetQueueName(),
		)
		return nil
	}

	// Declare queue
	if err := m.queueWrapper.DeclareMainQueue(worker.GetQueueName(), worker.IsRetryEnabled()); err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", worker.GetQueueName(), err)
	}

	// Declare DLQ if retry is disabled
	if !worker.IsRetryEnabled() {
		if err := m.queueWrapper.DeclareDLQ(worker.GetQueueName()); err != nil {
			return fmt.Errorf("failed to declare DLQ for %s: %w", worker.GetQueueName(), err)
		}
	} else {
		ttl := 30000 // default 30 detik (bisa dynamic nanti)
		if err := m.queueWrapper.DeclareRetryQueue(worker.GetQueueName(), ttl); err != nil {
			return fmt.Errorf("failed to declare retry queue for %s: %w", worker.GetQueueName(), err)
		}
	}

	m.workers = append(m.workers, worker)
	m.logger.Infof(
		"[+] Worker [%s] registered successfully (concurrency: %d, retry: %v, max_retry: %d)",
		worker.GetQueueName(),
		worker.GetConcurrency(),
		worker.IsRetryEnabled(),
		worker.GetMaxRetry(),
	)

	return nil
}

// Start begins consuming messages for all registered workers
func (m *WorkerManager) Start() error {
	for _, worker := range m.workers {
		// Start multiple consumers based on concurrency setting
		for i := 0; i < worker.GetConcurrency(); i++ {
			m.wg.Add(1)
			go m.startConsumer(worker, i+1)
		}
	}

	return nil
}

func (m *WorkerManager) startConsumer(worker Worker, consumerID int) {
	defer m.wg.Done()

	queueName := worker.GetQueueName()
	consumerTag := fmt.Sprintf("%s-consumer-%d", queueName, consumerID)

	for {
		select {
		case <-m.ctx.Done():
			// m.logger.Infof("Consumer [%s] shutting down", consumerTag)
			return
		default:
			// Attempt to start consuming with retry logic
			if err := m.consumeWithRetry(worker, consumerTag); err != nil {
				m.logger.WithError(err).Errorf(
					"Consumer [%s] stopped, retrying...",
					consumerTag,
				)

				select {
				case <-m.ctx.Done():
					return
				case <-time.After(5 * time.Second):
					// Retry
					continue
				}
			}
		}
	}
}

func (m *WorkerManager) consumeWithRetry(worker Worker, consumerTag string) error {
	queueName := worker.GetQueueName()

	// Wait for connection to be ready
	maxWait := 30 * time.Second
	waitInterval := 1 * time.Second
	totalWait := time.Duration(0)

	for !m.queueWrapper.IsConnected() {
		if totalWait >= maxWait {
			return fmt.Errorf("timeout waiting for RabbitMQ connection")
		}

		select {
		case <-m.ctx.Done():
			return fmt.Errorf("context cancelled while waiting for connection")
		case <-time.After(waitInterval):
			totalWait += waitInterval
		}
	}

	m.logger.Infof(
		"- Starting consumer [%s] for queue [%s]",
		consumerTag,
		queueName,
	)

	// Re-declare queue in case of reconnection
	if err := m.queueWrapper.DeclareMainQueue(queueName, worker.IsRetryEnabled()); err != nil {
		m.logger.WithError(err).Warnf(
			"Failed to re-declare queue [%s], will continue anyway",
			queueName,
		)
	}

	// Start consuming
	// deliveries, err := m.queueWrapper.Consume(queueName)
	// if err != nil {
	// 	return fmt.Errorf("failed to start consuming: %w", err)
	// }

	// Consumer channel (NEW)
	deliveries, ch, err := m.queueWrapper.ConsumeWithNewChannel(
		queueName,
		consumerTag,
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = ch.Cancel(consumerTag, false)
		_ = ch.Close()
	}()

	m.logger.Infof("- Consumer [%s] started successfully", consumerTag)

	// Process messages
	for {
		select {
		case <-m.ctx.Done():
			return nil

		case delivery, ok := <-deliveries:
			if !ok {
				m.logger.Warnf(
					"- Consumer [%s] channel closed, will reconnect...",
					consumerTag,
				)
				return fmt.Errorf("delivery channel closed")
			}

			m.processMessage(worker, delivery, consumerTag)
		}
	}
}

func (m *WorkerManager) processMessage(worker Worker, delivery amqp.Delivery, consumerTag string) {
	m.consumerWg.Add(1)
	defer m.consumerWg.Done()
	defer func() {
		if r := recover(); r != nil {
			m.logger.Errorf(
				"[%s] Panic recovered while processing message: %v",
				consumerTag,
				r,
			)
			// Try to reject the message
			delivery.Reject(false)
		}
	}()

	m.logger.Debugf(
		"[%s] Processing message: %s",
		consumerTag,
		delivery.MessageId,
	)

	// Process the message
	err := worker.Process(m.ctx, delivery)

	if err != nil {
		// Handle error
		worker.OnError(m.ctx, delivery, err)

		if worker.IsRetryEnabled() {
			var msg dto.MultiPostingWorkerMessage
			if jsonErr := json.Unmarshal(delivery.Body, &msg); jsonErr != nil {
				m.logger.WithError(jsonErr).Error("Failed to unmarshal message for retry check")
				_ = delivery.Reject(false)
				return
			}

			if msg.RetryCount >= msg.MaxRetry {
				m.logger.Warnf(
					"[%s] Retry exhausted for message %s, rejecting to DLQ",
					consumerTag,
					msg.MessageID,
				)
				_ = delivery.Reject(false)
				return
			}

			// retry masih ada → ACK
			if ackErr := delivery.Ack(false); ackErr != nil {
				m.logger.WithError(ackErr).Error("Failed to ack after retry publish")
			}
			return
		}

		// Reject message (will go to DLQ if configured)
		if rejectErr := delivery.Reject(false); rejectErr != nil {
			m.logger.WithError(rejectErr).Error(
				"Failed to reject message",
			)
		}
	} else {
		// Handle success
		worker.OnSuccess(m.ctx, delivery)

		// Acknowledge message
		if ackErr := delivery.Ack(false); ackErr != nil {
			m.logger.WithError(ackErr).Error(
				"Failed to acknowledge message",
			)
		}
	}
}

// Shutdown gracefully stops all workers
func (m *WorkerManager) Shutdown() {
	m.logger.Info("Shutting down worker manager...")

	// Signal all goroutines to stop
	m.cancel()

	// Wait for all consumers to stop
	m.wg.Wait()

	// Wait for all message processing to complete
	m.logger.Info("Waiting for message processing to complete...")
	m.consumerWg.Wait()

	m.logger.Info("Worker manager shut down successfully")
}

// GetWorkerCount returns the number of registered workers
func (m *WorkerManager) GetWorkerCount() int {
	return len(m.workers)
}
