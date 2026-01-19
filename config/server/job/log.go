package job

import (
	"background-job-service/config/logger"
	entity "background-job-service/internal/entity/worker"
	repositoryworker "background-job-service/internal/repository/worker"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// JobLogger handles database logging for scheduled jobs
type JobLogger struct {
	logger     *logger.Logger
	repository repositoryworker.LogWorkerRepositoryInterface
}

func NewJobLogger(
	logger *logger.Logger,
	repository repositoryworker.LogWorkerRepositoryInterface,
) *JobLogger {
	return &JobLogger{
		logger:     logger,
		repository: repository,
	}
}

// LogJobStart logs when a scheduled job starts
func (jl *JobLogger) LogJobStart(ctx context.Context, opt LogJobOption) {
	workerType := entity.WorkerTypeScheduledJob
	log := &entity.LogWorker{
		WorkerName:  opt.JobName,
		WorkerType:  &workerType,
		ReferenceID: opt.ReferenceID,
		Level:       entity.LogLevelInfo,
		Message:     fmt.Sprintf("Scheduled job '%s' started", opt.JobName),
		Context:     opt.Context,
	}

	go func() {
		if err := jl.repository.Create(context.Background(), log); err != nil {
			jl.logger.Logger.WithError(err).WithFields(logrus.Fields{
				"job_name": opt.JobName,
			}).Error("Failed to save job start log to database")
		}
	}()
}

// LogJobSuccess logs successful job execution
func (jl *JobLogger) LogJobSuccess(ctx context.Context, opt LogJobOption) {
	workerType := entity.WorkerTypeScheduledJob
	log := &entity.LogWorker{
		WorkerName:  opt.JobName,
		WorkerType:  &workerType,
		ReferenceID: opt.ReferenceID,
		Level:       entity.LogLevelInfo,
		Message:     opt.Message,
		Context:     opt.Context,
	}

	go func() {
		if err := jl.repository.Create(context.Background(), log); err != nil {
			jl.logger.Logger.WithError(err).WithFields(logrus.Fields{
				"job_name": opt.JobName,
			}).Error("Failed to save job success log to database")
		}
	}()
}

// LogJobError logs failed job execution
func (jl *JobLogger) LogJobError(ctx context.Context, opt LogJobOption, err error) {
	workerType := entity.WorkerTypeScheduledJob
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	log := &entity.LogWorker{
		WorkerName:  opt.JobName,
		WorkerType:  &workerType,
		ReferenceID: opt.ReferenceID,
		Level:       entity.LogLevelError,
		Message:     opt.Message,
		Error:       &errMsg,
		Context:     opt.Context,
	}

	go func() {
		if err := jl.repository.Create(context.Background(), log); err != nil {
			jl.logger.Logger.WithError(err).WithFields(logrus.Fields{
				"job_name": opt.JobName,
			}).Error("Failed to save job error log to database")
		}
	}()
}

type LogJobOption struct {
	JobName     string
	ReferenceID *string
	Message     string
	Context     json.RawMessage // Changed from *entity.JSONMap
}
