package rabbitmq

import (
	gommonlog "background-job-service/config/gommon"
	"background-job-service/pkg/constant"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type CircuitState int

const (
	CircuitClosed CircuitState = iota // normal
	CircuitOpen                       // stop reconnect
)

type RabbitMQWrapper struct {
	rabbitmq            *RabbitMQ
	channel             *amqp.Channel
	logger              *gommonlog.Logger
	conn                *amqp.Connection
	mu                  sync.RWMutex
	isConnected         bool
	reconnectRunning    bool
	closeChan           chan struct{}
	notifyClose         chan *amqp.Error
	notifyChannelClose  chan *amqp.Error
	healthCheckInterval time.Duration
	reconnectMaxRetries uint8
	reconnectInterval   time.Duration
	monitorWg           sync.WaitGroup

	// circuit breaker
	circuitState        CircuitState
	consecutiveFailures uint8
	maxFailures         uint8
	circuitOpenedAt     time.Time
	cooldownDuration    time.Duration
}

func NewRabbitMQWrapper(rmq *RabbitMQ, logger *gommonlog.Logger) (*RabbitMQWrapper, error) {
	wrapper := &RabbitMQWrapper{
		rabbitmq:            rmq,
		logger:              logger,
		closeChan:           make(chan struct{}),
		healthCheckInterval: time.Duration(rmq.Connection.Config.HealthCheckInterval) * time.Second,
		reconnectMaxRetries: rmq.Connection.Config.ReconnectMaxRetries,
		reconnectInterval:   time.Duration(rmq.Connection.Config.ReconnectInterval) * time.Second,

		// circuit breaker default
		circuitState:     CircuitClosed,
		maxFailures:      rmq.Connection.Config.CircuitBreakerMaxFailures,
		cooldownDuration: time.Duration(rmq.Connection.Config.CircuitBreakerTimeout) * time.Second,
	}

	// Initial connection
	if err := wrapper.connect(); err != nil {
		return nil, fmt.Errorf("failed initial connection to RabbitMQ: %w", err)
	}

	// Start health check
	go wrapper.startHealthCheck()

	return wrapper, nil
}

func (r *RabbitMQWrapper) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Close existing connections if any
	if r.channel != nil {
		r.channel.Close()
		r.channel = nil
	}
	if r.conn != nil {
		r.conn.Close()
		r.conn = nil
	}

	// Close old notification channels
	if r.notifyClose != nil {
		select {
		case <-r.notifyClose:
		default:
		}
		r.notifyClose = nil
	}
	if r.notifyChannelClose != nil {
		select {
		case <-r.notifyChannelClose:
		default:
		}
		r.notifyChannelClose = nil
	}

	conn, err := r.rabbitmq.Connect()
	if err != nil {
		r.isConnected = false
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		r.isConnected = false
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	if err := channel.Qos(1, 0, false); err != nil {
		channel.Close()
		conn.Close()
		r.isConnected = false
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	r.conn = conn
	r.channel = channel
	r.isConnected = true

	// Setup NEW notification channels
	r.notifyClose = make(chan *amqp.Error, 1)
	r.notifyChannelClose = make(chan *amqp.Error, 1)
	r.conn.NotifyClose(r.notifyClose)
	r.channel.NotifyClose(r.notifyChannelClose)

	r.logger.Infof("Connected to RabbitMQ at %v:%v", r.rabbitmq.Connection.DSN.Host, r.rabbitmq.Connection.DSN.Port)

	// Start NEW connection monitor - increment WaitGroup
	r.monitorWg.Add(1)
	go r.startConnectionMonitor()

	return nil
}

func (r *RabbitMQWrapper) startConnectionMonitor() {
	defer r.monitorWg.Done()

	for {
		select {
		case <-r.closeChan:
			r.logger.Info("Connection monitor stopped")
			return
		case err, ok := <-r.notifyClose:
			if !ok {
				// Channel closed normally, exit monitor
				return
			}
			if err != nil {
				r.logger.WithError(err).Error("RabbitMQ connection closed, attempting reconnect...")
				r.handleDisconnect()
				return // Exit and wait for new monitor after reconnect
			}
		case err, ok := <-r.notifyChannelClose:
			if !ok {
				// Channel closed normally, exit monitor
				return
			}
			if err != nil {
				r.logger.WithError(err).Error("RabbitMQ channel closed, attempting reconnect...")
				r.handleDisconnect()
				return // Exit and wait for new monitor after reconnect
			}
		}
	}
}

func (r *RabbitMQWrapper) handleDisconnect() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.circuitState == CircuitOpen {
		r.logger.Warn("🔴 Circuit breaker OPEN - skipping reconnect")
		return
	}

	if r.reconnectRunning {
		return
	}

	r.reconnectRunning = true
	r.isConnected = false

	go r.reconnect()
}

func (r *RabbitMQWrapper) reconnect() {
	defer func() {
		r.mu.Lock()
		r.reconnectRunning = false
		r.mu.Unlock()
	}()

	for i := 1; i <= int(r.reconnectMaxRetries); i++ {
		select {
		case <-r.closeChan:
			return
		default:
			r.logger.Infof("Reconnecting to RabbitMQ (attempt %d/%d)...", i, r.reconnectMaxRetries)

			if err := r.connect(); err != nil {
				r.logger.WithError(err).Error("Reconnect failed")
				time.Sleep(r.reconnectInterval)
				continue
			}

			// SUCCESS → reset breaker
			r.mu.Lock()
			r.consecutiveFailures = 0
			r.circuitState = CircuitClosed
			r.mu.Unlock()

			r.logger.Info("RabbitMQ reconnected successfully")
			return
		}
	}

	// FAILED ALL ATTEMPTS → open circuit
	r.openCircuit()
}

func (r *RabbitMQWrapper) openCircuit() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.consecutiveFailures++
	if r.consecutiveFailures < r.maxFailures {
		r.logger.Warnf(
			"Reconnect failed (%d/%d), will retry later",
			r.consecutiveFailures,
			r.maxFailures,
		)
		return
	}

	r.logger.Warnf(
		"Reconnect failed (%d/%d), will trigger Circuit breaker...",
		r.consecutiveFailures,
		r.maxFailures,
	)

	r.circuitState = CircuitOpen
	r.circuitOpenedAt = time.Now()

	r.logger.Warn("🔴 Circuit breaker OPEN - RabbitMQ reconnect paused")
}

func (r *RabbitMQWrapper) startHealthCheck() {
	ticker := time.NewTicker(r.healthCheckInterval)
	defer ticker.Stop()

	r.logger.Infof("RabbitMQ health check started (interval: %v)", r.healthCheckInterval)

	for {
		select {
		case <-r.closeChan:
			r.logger.Info("Health check stopped")
			return
		case <-ticker.C:
			r.performHealthCheck()
		}
	}
}

func (r *RabbitMQWrapper) performHealthCheck() {
	r.mu.RLock()
	state := r.circuitState
	openedAt := r.circuitOpenedAt
	cooldown := r.cooldownDuration
	isConnected := r.isConnected
	reconnectRunning := r.reconnectRunning
	channel := r.channel
	conn := r.conn
	r.mu.RUnlock()

	if state == CircuitOpen {
		if time.Since(openedAt) < cooldown {
			r.logger.Warn("🔴 Circuit breaker OPEN - waiting cooldown")
			return
		}

		r.logger.Info("🟡 Circuit breaker cooldown passed, trying reconnect")
		r.mu.Lock()
		r.circuitState = CircuitClosed
		r.consecutiveFailures = 0
		r.mu.Unlock()

		r.handleDisconnect()
		return
	}

	// Skip health check if already reconnecting
	if reconnectRunning {
		r.logger.Debug("Health check skipped (reconnecting in progress)")
		return
	}

	// If not connected, trigger reconnect
	if !isConnected {
		r.logger.Warn("Health check detected disconnection, attempting reconnect...")
		r.handleDisconnect()
		return
	}

	// Check if channel/connection is nil
	if channel == nil || conn == nil {
		r.logger.Warn("Health check: channel or connection is nil")
		r.handleDisconnect()
		return
	}

	// Check if connection is closed
	if conn.IsClosed() {
		r.logger.Warn("Health check: connection is closed")
		r.handleDisconnect()
		return
	}

	// SIMPLIFIED health check - just check if connection is alive
	// DON'T do QueueDeclarePassive as it can trigger channel errors
	r.logger.Infof("Health check RabbitMQ OK: %v:%v", r.rabbitmq.Connection.DSN.Host, r.rabbitmq.Connection.DSN.Port)
}

func (r *RabbitMQWrapper) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isConnected
}

func (r *RabbitMQWrapper) DeclareMainQueue(queueName string, enableRetry bool) error {
	r.mu.RLock()
	channel := r.channel
	isConnected := r.isConnected
	r.mu.RUnlock()

	if !isConnected || channel == nil {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	args := amqp.Table{}

	if !enableRetry {
		// DLQ HANYA jika retry tidak aktif
		args["x-dead-letter-exchange"] = "dlx"
		args["x-dead-letter-routing-key"] =
			fmt.Sprintf(constant.WORKER_DLQ_NAME, queueName)
	}

	_, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		// amqp.Table{
		// 	"x-message-ttl":             86400000, // 24 hours in milliseconds
		// 	"x-dead-letter-exchange":    "dlx",
		// 	"x-dead-letter-routing-key": fmt.Sprintf("%s.dlq", queueName),
		// },
		args,
	)
	if err != nil {
		r.logger.WithError(err).Warnf("Failed to declare queue [%s]", queueName)
	}
	return err
}

func (r *RabbitMQWrapper) DeclareRetryQueue(queueName string, ttlMs int) error {
	r.mu.RLock()
	ch := r.channel
	isConnected := r.isConnected
	r.mu.RUnlock()

	if !isConnected || ch == nil {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	retryQueue := fmt.Sprintf(constant.WORKER_RETRY_QUEUE_DECLARE_NAME, queueName)

	_, err := ch.QueueDeclare(
		retryQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl":             ttlMs,
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": queueName,
		},
	)

	return err
}

func (r *RabbitMQWrapper) DeclareDLQ(queueName string) error {
	r.mu.RLock()
	channel := r.channel
	isConnected := r.isConnected
	r.mu.RUnlock()

	if !isConnected || channel == nil {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	// Declare Dead Letter Exchange
	if err := channel.ExchangeDeclare(
		"dlx",    // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	); err != nil {
		return err
	}

	// Declare Dead Letter Queue
	dlqName := fmt.Sprintf(constant.WORKER_DLQ_NAME, queueName)
	_, err := channel.QueueDeclare(
		dlqName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	// Bind DLQ to DLX
	return channel.QueueBind(
		dlqName, // queue name
		dlqName, // routing key
		"dlx",   // exchange
		false,   // no-wait
		nil,     // arguments
	)
}

func (r *RabbitMQWrapper) Publish(queueName string, body []byte, priority uint8) error {
	r.mu.RLock()
	channel := r.channel
	isConnected := r.isConnected
	r.mu.RUnlock()

	if !isConnected || channel == nil {
		return fmt.Errorf("not connected to RabbitMQ")
	}

	return channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Priority:     priority,
			Timestamp:    time.Now(),
		},
	)
}

func (r *RabbitMQWrapper) Consume(queueName string) (<-chan amqp.Delivery, error) {
	r.mu.RLock()
	channel := r.channel
	isConnected := r.isConnected
	r.mu.RUnlock()

	if !isConnected || channel == nil {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	return channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (set to false for manual ack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
}

func (r *RabbitMQWrapper) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	close(r.closeChan)

	// Wait for all monitors to stop
	r.monitorWg.Wait()

	if r.channel != nil {
		r.channel.Close()
	}

	if r.conn != nil {
		r.conn.Close()
	}

	r.isConnected = false
	r.logger.Info("RabbitMQ connection closed")
	return nil
}

func (r *RabbitMQWrapper) GetChannel() *amqp.Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channel
}

func (r *RabbitMQWrapper) GetAMQPConnection() *amqp.Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn
}

func (r *RabbitMQWrapper) CreateConsumerChannel() (*amqp.Channel, error) {
	r.mu.RLock()
	conn := r.conn
	isConnected := r.isConnected
	r.mu.RUnlock()

	if !isConnected || conn == nil {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// QoS PER CONSUMER
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		return nil, err
	}

	return ch, nil
}

func (r *RabbitMQWrapper) ConsumeWithNewChannel(
	queueName string,
	consumerTag string,
) (<-chan amqp.Delivery, *amqp.Channel, error) {

	ch, err := r.CreateConsumerChannel()
	if err != nil {
		return nil, nil, err
	}

	deliveries, err := ch.Consume(
		queueName,   // queue
		consumerTag, // consumer
		false,       // auto-ack (set to false for manual ack)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		ch.Close()
		return nil, nil, err
	}

	return deliveries, ch, nil
}
