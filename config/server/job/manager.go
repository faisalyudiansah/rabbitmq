package job

import (
	gommonlog "background-job-service/config/gommon"
	"context"

	"github.com/go-co-op/gocron/v2"
)

// JobManager manages all scheduled jobs
type JobManager struct {
	scheduler gocron.Scheduler
	logger    *gommonlog.Logger
	jobs      []ScheduledJob
}

func NewJobManager(logger *gommonlog.Logger, scheduler gocron.Scheduler) *JobManager {
	return &JobManager{
		logger:    logger,
		scheduler: scheduler,
		jobs:      make([]ScheduledJob, 0),
	}
}

// RegisterJob adds a job to the manager
func (m *JobManager) RegisterJob(job ScheduledJob) error {
	if !job.IsEnabled() {
		m.logger.Infof(
			"Scheduled Job [%s] is disabled, skipping registration",
			job.GetName(),
		)
		return nil
	}

	// Create the job task
	_, err := m.scheduler.NewJob(
		gocron.CronJob(job.GetSchedule(), true),
		gocron.NewTask(func() {
			ctx := context.Background()

			m.logger.Infof(
				"[Scheduled Job: %s] Starting execution...",
				job.GetName(),
			)

			if err := job.Execute(ctx); err != nil {
				job.OnError(ctx, err)
			} else {
				job.OnSuccess(ctx)
			}
		}),
		gocron.WithName(job.GetName()),
	)

	if err != nil {
		return err
	}

	m.jobs = append(m.jobs, job)
	m.logger.Infof(
		"- Scheduled Job [%s] registered successfully (schedule: %s)",
		job.GetName(),
		job.GetSchedule(),
	)

	return nil
}

// GetJobCount returns the number of registered jobs
func (m *JobManager) GetJobCount() int {
	return len(m.jobs)
}
