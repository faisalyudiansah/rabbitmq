package job

import (
	"background-job-service/config"
	"background-job-service/config/logger"
	"context"
)

// ScheduledJob interface that all scheduled jobs must implement
type ScheduledJob interface {
	// GetName returns the job name for logging
	GetName() string

	// GetSchedule returns the cron expression for this job
	GetSchedule() string

	// IsEnabled checks if this job is enabled via config
	IsEnabled() bool

	// Execute runs the actual job logic
	Execute(ctx context.Context) error

	// OnError is called when job execution fails
	OnError(ctx context.Context, err error)

	// OnSuccess is called when job execution succeeds
	OnSuccess(ctx context.Context)
}

// BaseJob provides common functionality
type BaseJob struct {
	Name     string
	Schedule string
	Logger   *logger.Logger
	Cfg      *config.Config
	Enabled  bool
}

func (b *BaseJob) GetName() string {
	return b.Name
}

func (b *BaseJob) GetSchedule() string {
	return b.Schedule
}

func (b *BaseJob) IsEnabled() bool {
	return b.Enabled
}

func (b *BaseJob) OnError(ctx context.Context, err error) {
	b.Logger.Logger.WithError(err).Errorf(
		"[Scheduled Job: %s] Execution failed",
		b.Name,
	)
}

func (b *BaseJob) OnSuccess(ctx context.Context) {
	b.Logger.Logger.Infof(
		"[Scheduled Job: %s] Execution completed successfully",
		b.Name,
	)
}
