package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type LogLevel string
type WorkerType string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

const (
	WorkerTypeRabbitMQ     WorkerType = "rabbitmq_worker"
	WorkerTypeScheduledJob WorkerType = "scheduled_job"
)

type LogWorker struct {
	ID         int64
	WorkerName string
	WorkerType *WorkerType

	BatchID     *string
	MessageID   *string
	ReferenceID *string

	Level LogLevel

	Message string
	Error   *string

	Context json.RawMessage

	CreatedAt time.Time
	CreatedBy *string
}

func (LogWorker) TableName() string {
	return "service_sales_return.log_workers"
}

type LogWorkerFilter struct {
	WorkerName  *string
	BatchID     *string
	ReferenceID *string
	MessageID   *string
	Level       *string
}

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
