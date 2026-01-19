package server

import (
	"background-job-service/config"
	gommonlog "background-job-service/config/gommon"
	"background-job-service/config/logger"
	"background-job-service/config/rabbitmq"
	"background-job-service/config/server/job"
	"background-job-service/config/server/worker"

	"github.com/go-co-op/gocron/v2"
)

type BootstrapWorkerConfig struct {
	Cfg           *config.Config
	Scheduler     gocron.Scheduler
	Log           *gommonlog.Logger
	Logger        *logger.Logger
	QueueWrapper  *rabbitmq.RabbitMQWrapper
	WorkerManager *worker.WorkerManager
	JobManager    *job.JobManager
}

// Bootstrap your worker here
func (c *BootstrapWorkerConfig) Bootstrap() error {
	// Register Scheduled Jobs
	if err := c.registerScheduledJobs(); err != nil {
		return err
	}

	// Register RabbitMQ Workers
	if err := c.registerWorkers(); err != nil {
		return err
	}

	// Start consuming messages
	if err := c.WorkerManager.Start(); err != nil {
		return err
	}

	return nil
}

func (c *BootstrapWorkerConfig) registerWorkers() error {
	c.Log.Info("Registering RabbitMQ workers...")

	// workerLogger := worker.NewWorkerLogger(c.Logger, LogWorkerRepository)

	// // Register Multi Posting Worker
	// multiPostingWorker := worker.NewMultiPostingWorker(
	// 	c.Logger,
	// 	c.QueueWrapper,
	// 	workerLogger,
	// 	SalesReturnCNUseCase,
	// 	c.Config,
	// )
	// if err := c.WorkerManager.RegisterWorker(multiPostingWorker); err != nil {
	// 	return err
	// }

	// Future workers can be registered here
	// Example:
	// emailWorker := worker.NewEmailWorker(...)
	// if err := c.WorkerManager.RegisterWorker(emailWorker); err != nil {
	//     return err
	// }

	// notificationWorker := worker.NewNotificationWorker(...)
	// if err := c.WorkerManager.RegisterWorker(notificationWorker); err != nil {
	//     return err
	// }

	c.Log.Infof(
		"[ADD] %d worker(s) registered successfully",
		c.WorkerManager.GetWorkerCount(),
	)

	return nil
}

func (c *BootstrapWorkerConfig) registerScheduledJobs() error {
	c.Log.Info("Registering scheduled jobs...")

	// 1. Auto Close Stale Returns Job (with dependencies)
	// autoCloseStaleReturnsJob := job.NewAutoCloseStaleReturnsJob(
	// 	job.AutoCloseStaleReturnsJobOption{
	// 		Logger:        c.Logger,
	// 		Config:        c.Config,
	// 		NotaReturRepo: NotaReturRepo, // from global config
	// 		DaysThreshold: 30,
	// 	},
	// )
	// if err := c.JobManager.RegisterJob(autoCloseStaleReturnsJob); err != nil {
	// 	return err
	// }

	// 2. Daily Sales Return Summary Job
	// dailySummaryJob := job.NewDailySalesReturnSummaryJob(
	// 	job.DailySalesReturnSummaryJobOption{
	// 		Logger:            c.Logger,
	// 		Config:            c.Config,
	// 		NotaReturRepo:     NotaReturRepo,
	// 		NotaReturItemRepo: NotaReturItemRepo,
	// 	},
	// )
	// if err := c.JobManager.RegisterJob(dailySummaryJob); err != nil {
	// 	return err
	// }

	c.Log.Infof(
		"[ADD] %d scheduled job(s) registered successfully",
		c.JobManager.GetJobCount(),
	)

	return nil
}
