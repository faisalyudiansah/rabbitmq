package scheduler

import (
	"background-job-service/config/logger"
	"errors"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
)

var (
	scheduler     gocron.Scheduler
	schedulerOnce sync.Once
)

// SchedulerLogger implement gocron.Logger interface
type SchedulerLogger struct {
	Logger *logger.Logger
}

func (l *SchedulerLogger) Error(msg string, args ...interface{}) {
	l.Logger.Logger.WithError(errors.New(msg)).Error(args...)
}

func (l *SchedulerLogger) Warn(msg string, args ...interface{}) {
	l.Logger.Logger.Warnf(msg, args...)
}

func (l *SchedulerLogger) Info(msg string, args ...interface{}) {
	l.Logger.Logger.WithField("msg", msg).Info(args...)
}

func (l *SchedulerLogger) Debug(msg string, args ...interface{}) {
	l.Logger.Logger.WithField("msg", msg).Debug(args...)
}

type SchedulerOption struct {
	Timezone string
	Logger   *SchedulerLogger
}

// NewScheduler wraps gocron.Scheduler
func NewScheduler(s *SchedulerOption) (gocron.Scheduler, error) {
	var err error

	timezone, err := time.LoadLocation(s.Timezone)

	if err != nil {
		return nil, err
	}

	schedulerOnce.Do(func() {
		scheduler, err = gocron.NewScheduler(
			gocron.WithLocation(timezone),
			gocron.WithLogger(s.Logger),
		)
	})

	if err != nil {
		return nil, err
	}

	return scheduler, nil
}
