package worker

import (
	"background-job-service/config/logger"
	entity "background-job-service/internal/entity/worker"
	repositoryworker "background-job-service/internal/repository/worker"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// WorkerLogger handles database logging for workers
type WorkerLogger struct {
	logger     *logger.Logger
	repository repositoryworker.LogWorkerRepositoryInterface
}

func NewWorkerLogger(
	logger *logger.Logger,
	repository repositoryworker.LogWorkerRepositoryInterface,
) *WorkerLogger {
	return &WorkerLogger{
		logger:     logger,
		repository: repository,
	}
}

// LogSuccess logs successful worker execution to database
func (wl *WorkerLogger) LogSuccess(ctx context.Context, opt LogWorkerOption) {
	log := &entity.LogWorker{
		WorkerName:  opt.WorkerName,
		WorkerType:  opt.WorkerType,
		BatchID:     opt.BatchID,
		MessageID:   opt.MessageID,
		ReferenceID: opt.ReferenceID,
		Level:       entity.LogLevelInfo,
		Message:     opt.Message,
		Context:     opt.Context,
		CreatedBy:   opt.CreatedBy,
	}

	// Save to database asynchronously
	go func() {
		if err := wl.repository.Create(context.Background(), log); err != nil {
			wl.logger.Logger.WithError(err).WithFields(logrus.Fields{
				"worker_name": opt.WorkerName,
				"message_id":  opt.MessageID,
				"batch_id":    opt.BatchID,
			}).Error("Failed to save worker success log to database")
		}
	}()
}

// LogError logs failed worker execution to database
func (wl *WorkerLogger) LogError(ctx context.Context, opt LogWorkerOption, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	log := &entity.LogWorker{
		WorkerName:  opt.WorkerName,
		WorkerType:  opt.WorkerType,
		BatchID:     opt.BatchID,
		MessageID:   opt.MessageID,
		ReferenceID: opt.ReferenceID,
		Level:       entity.LogLevelError,
		Message:     opt.Message,
		Error:       &errMsg,
		Context:     opt.Context,
		CreatedBy:   opt.CreatedBy,
	}

	// Save to database asynchronously
	go func() {
		if err := wl.repository.Create(context.Background(), log); err != nil {
			wl.logger.Logger.WithError(err).WithFields(logrus.Fields{
				"worker_name": opt.WorkerName,
				"message_id":  opt.MessageID,
				"batch_id":    opt.BatchID,
			}).Error("Failed to save worker error log to database")
		}
	}()
}

// LogInfo logs informational messages
func (wl *WorkerLogger) LogInfo(ctx context.Context, opt LogWorkerOption) {
	log := &entity.LogWorker{
		WorkerName:  opt.WorkerName,
		WorkerType:  opt.WorkerType,
		BatchID:     opt.BatchID,
		MessageID:   opt.MessageID,
		ReferenceID: opt.ReferenceID,
		Level:       entity.LogLevelInfo,
		Message:     opt.Message,
		Context:     opt.Context,
		CreatedBy:   opt.CreatedBy,
	}

	go func() {
		if err := wl.repository.Create(context.Background(), log); err != nil {
			wl.logger.Logger.WithError(err).Error("Failed to save worker info log")
		}
	}()
}

// LogWarning logs warning messages
func (wl *WorkerLogger) LogWarning(ctx context.Context, opt LogWorkerOption) {
	log := &entity.LogWorker{
		WorkerName:  opt.WorkerName,
		WorkerType:  opt.WorkerType,
		BatchID:     opt.BatchID,
		MessageID:   opt.MessageID,
		ReferenceID: opt.ReferenceID,
		Level:       entity.LogLevelWarn,
		Message:     opt.Message,
		Context:     opt.Context,
		CreatedBy:   opt.CreatedBy,
	}

	go func() {
		if err := wl.repository.Create(context.Background(), log); err != nil {
			wl.logger.Logger.WithError(err).Error("Failed to save worker warning log")
		}
	}()
}

type LogWorkerOption struct {
	WorkerName  string
	WorkerType  *entity.WorkerType // "rabbitmq" or "scheduled_job"
	BatchID     *string
	MessageID   *string
	ReferenceID *string // Some ID, Order ID, etc
	Message     string
	Context     json.RawMessage
	CreatedBy   *string
}

// Helper to add additional context
func AddContextField(ctx json.RawMessage, key string, value interface{}) json.RawMessage {

	contextMap := make(map[string]interface{})

	if len(ctx) > 0 {
		if err := json.Unmarshal(ctx, &contextMap); err != nil {
			contextMap = make(map[string]interface{})
		}
	}

	contextMap[key] = value

	bytes, err := json.Marshal(contextMap)
	if err != nil {
		return ctx
	}

	return json.RawMessage(bytes)
}

// Helper to format success message
func FormatSuccessMessage(workerName, action string, details map[string]interface{}) string {
	msg := fmt.Sprintf("[%s] %s completed successfully", workerName, action)
	if len(details) > 0 {
		detailsBytes, _ := json.Marshal(details)
		msg += fmt.Sprintf(" - Details: %s", string(detailsBytes))
	}
	return msg
}

// Helper to format error message
func FormatErrorMessage(workerName, action string, err error) string {
	return fmt.Sprintf("[%s] %s failed: %v", workerName, action, err)
}

// Helper to create context from message body
func CreateContextFromMessage(messageBody interface{}) json.RawMessage {
	if messageBody == nil {
		return nil
	}

	bytes, err := json.Marshal(messageBody)
	if err != nil {
		return nil
	}

	return json.RawMessage(bytes)
}
