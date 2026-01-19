package job

// import (
// 	"background-job-service/config"
// 	"background-job-service/config/logger"
// 	"background-job-service/config/logstash"
// 	"background-job-service/config/rabbitmq"
// 	"background-job-service/pkg/constant"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/sirupsen/logrus"
// )

// // ============================================
// // Example: Scheduled Job that Publishes to RabbitMQ
// // ============================================
// type ScheduledMultiPostingJob struct {
// 	BaseJob
// 	NotaReturRepo *repository.NotaReturRepository
// 	QueueWrapper  *rabbitmq.RabbitMQWrapper
// 	MaxRetry      uint8
// }

// type ScheduledMultiPostingJobOption struct {
// 	Logger        *logger.Logger
// 	Config        *config.Config
// 	NotaReturRepo *repository.NotaReturRepository
// 	QueueWrapper  *rabbitmq.RabbitMQWrapper
// }

// func NewScheduledMultiPostingJob(opt ScheduledMultiPostingJobOption) *ScheduledMultiPostingJob {
// 	schedule := opt.Config.GetString(constant.JOB_SCHEDULED_MULTI_POSTING_SCHEDULE)
// 	if schedule == "" {
// 		schedule = "0 2 * * *" // 2 AM every day
// 	}

// 	maxRetry := opt.Config.GetInt(constant.WORKER_MULTI_POSTING_MAX_RETRY)
// 	if maxRetry == 0 {
// 		maxRetry = 3
// 	}

// 	return &ScheduledMultiPostingJob{
// 		BaseJob: BaseJob{
// 			Name:     "ScheduledMultiPostingJob",
// 			Schedule: schedule,
// 			Logger:   opt.Logger,
// 			Cfg:      opt.Config,
// 			Enabled:  opt.Config.GetBool(constant.JOB_SCHEDULED_MULTI_POSTING_ENABLED),
// 		},
// 		NotaReturRepo: opt.NotaReturRepo,
// 		QueueWrapper:  opt.QueueWrapper,
// 		MaxRetry:      maxRetry,
// 	}
// }

// func (j *ScheduledMultiPostingJob) Execute(ctx context.Context) error {
// 	j.Logger.Logger.Info("Starting scheduled multi posting job...")

// 	logstash.Info("SCHEDULED_MULTI_POSTING_START", logrus.Fields{
// 		"job_name": j.Name,
// 	})

// 	// TODO: Find CN that need to be posted
// 	// Example: Get all approved CN that haven't been posted yet
// 	// cns, err := j.NotaReturRepo.FindPendingCNForPosting(ctx)
// 	// if err != nil {
// 	//     return fmt.Errorf("failed to find pending CNs: %w", err)
// 	// }

// 	// For demonstration
// 	pendingCNs := []string{} // Replace with actual query result
// 	batchID := uuid.New().String()
// 	publishedCount := 0
// 	failedCount := 0

// 	j.Logger.Logger.WithFields(logrus.Fields{
// 		"batch_id":         batchID,
// 		"pending_cn_count": len(pendingCNs),
// 	}).Info("Found pending CNs for posting")

// 	// Publish messages to RabbitMQ with retry
// 	totalInBatch := len(pendingCNs)
// 	for i, cnID := range pendingCNs {
// 		message := dtopkg.MultiPostingWorkerMessage{
// 			WorkerMessage: dtopkg.WorkerMessage{
// 				MessageID:  uuid.New().String(),
// 				Type:       "multi_posting",
// 				CreatedAt:  time.Now(),
// 				RetryCount: 0,
// 				MaxRetry:   3, // Default, will be overridden by worker config
// 			},
// 			BatchID:      batchID,
// 			CNID:         cnID,
// 			PostingDate:  time.Now(),
// 			UpdatedBy:    "SYSTEM_SCHEDULED_JOB",
// 			TotalInBatch: totalInBatch,
// 			IndexInBatch: i + 1,
// 			// HeaderHTTPRequest: dtopkg.HeaderHTTPRequest{
// 			// 	CompanyID:  c.Request().Header.Get(constant.HEADER_X_COMPANY_ID),
// 			// 	Email:      c.Request().Header.Get(constant.HEADER_X_EMAIL),
// 			// 	OrgDesc:    c.Request().Header.Get(constant.HEADER_X_ORG_DESC),
// 			// 	SalesOrgID: c.Request().Header.Get(constant.HEADER_X_SALES_ORG_ID),
// 			// },
// 		}

// 		if err := j.publishWithRetry(message, 3); err != nil {
// 			j.Logger.Logger.WithError(err).WithFields(logrus.Fields{
// 				"cn_id":      cnID,
// 				"message_id": message.MessageID,
// 			}).Error("Failed to publish message after retries")

// 			logstash.ErrorSync("SCHEDULED_MULTI_POSTING_PUBLISH_FAILED", logrus.Fields{
// 				"job_name":   j.Name,
// 				"cn_id":      cnID,
// 				"message_id": message.MessageID,
// 				"batch_id":   batchID,
// 				"error":      err.Error(),
// 			})

// 			failedCount++
// 			continue
// 		}

// 		publishedCount++
// 	}

// 	j.Logger.Logger.WithFields(logrus.Fields{
// 		"batch_id":        batchID,
// 		"published_count": publishedCount,
// 		"failed_count":    failedCount,
// 	}).Info("Scheduled multi posting job completed")

// 	logstash.Info("SCHEDULED_MULTI_POSTING_END", logrus.Fields{
// 		"job_name":        j.Name,
// 		"batch_id":        batchID,
// 		"published_count": publishedCount,
// 		"failed_count":    failedCount,
// 		"status":          "completed",
// 	})

// 	if failedCount > 0 {
// 		return fmt.Errorf("completed with %d failures out of %d total", failedCount, len(pendingCNs))
// 	}

// 	return nil
// }

// // publishWithRetry attempts to publish message with retry logic
// func (j *ScheduledMultiPostingJob) publishWithRetry(message dtopkg.MultiPostingWorkerMessage, maxAttempts int) error {
// 	messageBytes, err := json.Marshal(message)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal message: %w", err)
// 	}

// 	var lastErr error
// 	for attempt := 1; attempt <= maxAttempts; attempt++ {
// 		// Check if RabbitMQ is connected
// 		if !j.QueueWrapper.IsConnected() {
// 			j.Logger.Logger.Warnf(
// 				"RabbitMQ not connected, waiting (attempt %d/%d)...",
// 				attempt,
// 				maxAttempts,
// 			)
// 			time.Sleep(time.Duration(attempt) * time.Second)
// 			continue
// 		}

// 		// Try to publish
// 		err := j.QueueWrapper.Publish(constant.WORKER_QUEUE_MULTI_POSTING, messageBytes, 5)
// 		if err == nil {
// 			if attempt > 1 {
// 				j.Logger.Logger.Infof(
// 					"Successfully published message after %d attempts",
// 					attempt,
// 				)
// 			}
// 			return nil
// 		}

// 		lastErr = err
// 		j.Logger.Logger.WithError(err).Warnf(
// 			"Failed to publish message (attempt %d/%d)",
// 			attempt,
// 			maxAttempts,
// 		)

// 		// Wait before retry with exponential backoff
// 		if attempt < maxAttempts {
// 			backoff := time.Duration(attempt*attempt) * time.Second
// 			time.Sleep(backoff)
// 		}
// 	}

// 	return fmt.Errorf("failed to publish after %d attempts: %w", maxAttempts, lastErr)
// }

// // ============================================
// // Example: Auto-Close Stale Returns Job
// // (This job doesn't need RabbitMQ)
// // ============================================
// type AutoCloseStaleReturnsJob struct {
// 	BaseJob
// 	NotaReturRepo *repository.NotaReturRepository
// 	DaysThreshold int
// }

// type AutoCloseStaleReturnsJobOption struct {
// 	Logger        *logger.Logger
// 	Config        *config.Config
// 	NotaReturRepo *repository.NotaReturRepository
// 	DaysThreshold uint8
// }

// func NewAutoCloseStaleReturnsJob(opt AutoCloseStaleReturnsJobOption) *AutoCloseStaleReturnsJob {
// 	schedule := opt.Config.GetString(constant.JOB_AUTO_CLOSE_STALE_RETURNS_SCHEDULE)
// 	if schedule == "" {
// 		schedule = constant.CRON_DAILY_MIDNIGHT
// 	}

// 	daysThreshold := opt.DaysThreshold
// 	if daysThreshold <= 0 {
// 		daysThreshold = 30
// 	}

// 	return &AutoCloseStaleReturnsJob{
// 		BaseJob: BaseJob{
// 			Name:     "AutoCloseStaleReturnsJob",
// 			Schedule: schedule,
// 			Logger:   opt.Logger,
// 			Config:   opt.Config,
// 			Enabled:  opt.Config.GetBool(constant.JOB_AUTO_CLOSE_STALE_RETURNS_ENABLED),
// 		},
// 		NotaReturRepo: opt.NotaReturRepo,
// 		DaysThreshold: daysThreshold,
// 	}
// }

// func (j *AutoCloseStaleReturnsJob) Execute(ctx context.Context) error {
// 	j.Logger.Logger.WithFields(logrus.Fields{
// 		"days_threshold": j.DaysThreshold,
// 	}).Info("Starting auto-close stale returns job...")

// 	logstash.Info("AUTO_CLOSE_STALE_RETURNS_START", logrus.Fields{
// 		"job_name":       j.Name,
// 		"days_threshold": j.DaysThreshold,
// 	})

// 	cutoffDate := time.Now().AddDate(0, 0, -j.DaysThreshold)

// 	// No RabbitMQ needed - just database operations
// 	closedCount := 0
// 	// Implementation here...

// 	j.Logger.Logger.WithFields(logrus.Fields{
// 		"closed_count": closedCount,
// 		"cutoff_date":  cutoffDate.Format("2006-01-02"),
// 	}).Info("Auto-close stale returns completed")

// 	logstash.Info("AUTO_CLOSE_STALE_RETURNS_END", logrus.Fields{
// 		"job_name":     j.Name,
// 		"closed_count": closedCount,
// 		"cutoff_date":  cutoffDate.Format("2006-01-02"),
// 		"status":       "success",
// 	})

// 	return nil
// }

// // ============================================
// // Example: Generate Daily Sales Return Summary
// // ============================================
// type DailySalesReturnSummaryJob struct {
// 	BaseJob
// 	NotaReturRepo     *repository.NotaReturRepository
// 	NotaReturItemRepo *repository.NotaReturItemRepository
// }

// type DailySalesReturnSummaryJobOption struct {
// 	Logger            *logger.Logger
// 	Config            *config.Config
// 	NotaReturRepo     *repository.NotaReturRepository
// 	NotaReturItemRepo *repository.NotaReturItemRepository
// }

// func NewDailySalesReturnSummaryJob(opt DailySalesReturnSummaryJobOption) *DailySalesReturnSummaryJob {
// 	schedule := opt.Config.GetString(constant.JOB_DAILY_SALES_RETURN_SUMMARY_SCHEDULE)
// 	if schedule == "" {
// 		schedule = "0 6 * * *" // 6 AM every day
// 	}

// 	return &DailySalesReturnSummaryJob{
// 		BaseJob: BaseJob{
// 			Name:     "DailySalesReturnSummaryJob",
// 			Schedule: schedule,
// 			Logger:   opt.Logger,
// 			Config:   opt.Config,
// 			Enabled:  opt.Config.GetBool(constant.JOB_DAILY_SALES_RETURN_SUMMARY_ENABLED),
// 		},
// 		NotaReturRepo:     opt.NotaReturRepo,
// 		NotaReturItemRepo: opt.NotaReturItemRepo,
// 	}
// }

// func (j *DailySalesReturnSummaryJob) Execute(ctx context.Context) error {
// 	j.Logger.Logger.Info("Generating daily sales return summary...")

// 	logstash.Info("DAILY_SALES_RETURN_SUMMARY_START", logrus.Fields{
// 		"job_name": j.Name,
// 	})

// 	// Get yesterday's date range
// 	now := time.Now()
// 	yesterday := now.AddDate(0, 0, -1)
// 	startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
// 	endOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, yesterday.Location())

// 	j.Logger.Logger.WithFields(logrus.Fields{
// 		"start_date": startOfDay.Format("2006-01-02 15:04:05"),
// 		"end_date":   endOfDay.Format("2006-01-02 15:04:05"),
// 	}).Info("Fetching returns for date range")

// 	// TODO: Implement summary generation logic
// 	// Example:
// 	// 1. Count total returns created yesterday
// 	// 2. Calculate total return amount
// 	// 3. Group by warehouse/sales org
// 	// 4. Generate report
// 	// 5. Send email or save to database

// 	summary := map[string]interface{}{
// 		"date":                 yesterday.Format("2006-01-02"),
// 		"total_returns":        0,
// 		"total_amount":         0.0,
// 		"returns_by_warehouse": map[string]int{},
// 		"top_returned_items":   []string{},
// 	}

// 	j.Logger.Logger.WithFields(logrus.Fields{
// 		"summary": fmt.Sprintf("%+v", summary),
// 	}).Info("Daily sales return summary generated")

// 	logstash.Info("DAILY_SALES_RETURN_SUMMARY_END", logrus.Fields{
// 		"job_name": j.Name,
// 		"date":     yesterday.Format("2006-01-02"),
// 		"summary":  summary,
// 		"status":   "success",
// 	})

// 	return nil
// }
