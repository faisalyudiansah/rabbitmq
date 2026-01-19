package main

import (
	"background-job-service/config"
	gommonlog "background-job-service/config/gommon"
	"background-job-service/config/logger"
	"background-job-service/config/logstash"
	"background-job-service/config/rabbitmq"
	"background-job-service/config/scheduler"
	"background-job-service/config/server"
	"background-job-service/config/server/job"
	"background-job-service/config/server/worker"
	"background-job-service/pkg/constant"
	"context"
	"crypto/tls"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	loggerConfig := logger.LoggerConfig{
		OutputDir: constant.LOG_LOCATION_FILE,
		FileName:  constant.LOG_FILE_NAME_WORKER,
	}

	logging, err := logger.NewLogger(logrus.New(), loggerConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Logger initialization failed"))
	}

	log := gommonlog.New()
	cfg := config.LoadConfig()

	logstash.InitLogstash(cfg, log)

	rabbitmqConfig := rabbitmq.NewRabbitMQConfig(cfg, logging, &tls.Config{})

	rabbitMQWrapper, err := rabbitmq.NewRabbitMQWrapper(rabbitmqConfig, log)
	if err != nil {
		logstash.ErrorSync("RABBITMQ_WRAPPER", logrus.Fields{
			"error": err.Error(),
		})
		log.Fatalf("Failed to create RabbitMQ wrapper: %v", err)
	}

	logging.Logger.WithFields(logrus.Fields{
		"health_check_interval":        cfg.RabbitMQ.HealthCheckInterval,
		"reconnect_max_retries":        cfg.RabbitMQ.ReconnectMaxRetries,
		"reconnect_interval":           cfg.RabbitMQ.ReconnectInterval,
		"circuit_breaker_max_failures": cfg.RabbitMQ.CircuitBreakerMaxFailures,
		"circuit_breaker_timeout":      cfg.RabbitMQ.CircuitBreakerTimeout,
	}).Info("RabbitMQ wrapper initialized with auto-reconnect and health check...")

	workerScheduler, err := scheduler.NewScheduler(&scheduler.SchedulerOption{
		Timezone: cfg.Timezone,
		Logger: &scheduler.SchedulerLogger{
			Logger: logging,
		},
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "Scheduler initialization failed"))
	}

	// Initialize Worker Manager
	workerManager := worker.NewWorkerManager(log, rabbitMQWrapper)

	// Initialize Job Manager
	jobManager := job.NewJobManager(log, workerScheduler)

	bootstrapConfigWorker := &server.BootstrapWorkerConfig{
		Scheduler:     workerScheduler,
		Cfg:           cfg,
		Log:           log,
		Logger:        logging,
		QueueWrapper:  rabbitMQWrapper,
		WorkerManager: workerManager,
		JobManager:    jobManager,
	}

	if err := bootstrapConfigWorker.Bootstrap(); err != nil {
		logstash.ErrorSync("WORKER_BOOTSTRAP", logrus.Fields{
			"error": err.Error(),
		})
		log.WithError(err).Fatal("Worker bootstrap failed")
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	// Start scheduler
	log.Info("Starting gocron scheduler...")
	workerScheduler.Start()

	log.Info("Worker service is running with auto-reconnect enabled")

	// Wait for shutdown signal
	<-ctx.Done()
	log.Warn("Shutdown signal received, initiating graceful shutdown...")

	// Graceful shutdown with timeout

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.TimeoutGracefulShutdown)*time.Second,
	)
	defer cancel()

	// Shutdown components
	done := make(chan struct{})
	go func() {
		// Shutdown scheduler
		log.Info("Shutting down gocron scheduler...")
		if err := workerScheduler.Shutdown(); err != nil {
			logstash.ErrorSync("SCHEDULER_SHUTDOWN", logrus.Fields{
				"error": err.Error(),
			})
			log.WithError(err).Error("Error while shutting down scheduler")
		}

		// Shutdown worker manager (stops consuming new messages)
		workerManager.Shutdown()

		// Close RabbitMQ wrapper
		log.Info("Closing RabbitMQ connection...")
		if err := rabbitMQWrapper.Close(); err != nil {
			log.WithError(err).Error("Error closing RabbitMQ wrapper")
		}

		close(done)
	}()

	select {
	case <-done:
		log.Info("All components shut down successfully")
	case <-shutdownCtx.Done():
		log.Warn("Shutdown timeout reached, forcing exit")
	}

	log.Info("Worker service stopped gracefully")
}
