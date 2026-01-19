package worker

// import (
// 	"background-job-service/config"
// 	"background-job-service/config/logger"
// 	"background-job-service/config/logstash"
// 	"background-job-service/config/rabbitmq"
// 	entity "background-job-service/internal/entity/worker"
// 	"background-job-service/pkg/constant"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"

// 	"github.com/labstack/echo/v4"
// 	amqp "github.com/rabbitmq/amqp091-go"
// 	"github.com/sirupsen/logrus"
// )

// type MultiPostingWorker struct {
// 	BaseWorker
// 	UseCase usecase.SalesReturnCNUseCaseInterface
// 	Config  *config.Config
// }

// func NewMultiPostingWorker(
// 	logger *logger.Logger,
// 	queueWrapper *rabbitmq.RabbitMQWrapper,
// 	workerLogger *WorkerLogger,
// 	useCase usecase.SalesReturnCNUseCaseInterface,
// 	config *config.Config,
// ) *MultiPostingWorker {
// 	return &MultiPostingWorker{
// 		BaseWorker: BaseWorker{
// 			QueueName:    constant.WORKER_QUEUE_MULTI_POSTING,
// 			Logger:       logger,
// 			QueueWrapper: queueWrapper,
// 			WorkerLogger: workerLogger,
// 			Enabled:      config.MultiPostingWorker.Enabled,
// 			RetryEnabled: config.MultiPostingWorker.RetryEnabled,
// 			MaxRetry:     config.MultiPostingWorker.MaxRetry,
// 			Concurrency:  config.MultiPostingWorker.Concurrency,
// 		},
// 		UseCase: useCase,
// 		Config:  config,
// 	}
// }

// func (w *MultiPostingWorker) Process(ctx context.Context, delivery amqp.Delivery) error {
// 	// Parse message
// 	var message dtopkg.MultiPostingWorkerMessage
// 	if err := json.Unmarshal(delivery.Body, &message); err != nil {
// 		w.Logger.Logger.WithError(err).Error("Failed to unmarshal multi posting message")
// 		return fmt.Errorf("invalid message format: %w", err)
// 	}

// 	w.Logger.Logger.WithFields(logrus.Fields{
// 		"message_id":     message.MessageID,
// 		"cn_id":          message.CNID,
// 		"batch_id":       message.BatchID,
// 		"index_in_batch": message.IndexInBatch,
// 		"retry_count":    message.RetryCount,
// 		"enable_retry":   w.IsRetryEnabled(),
// 	}).Info("Processing multi posting message")

// 	// Log to logstash
// 	logstash.Info("MULTI_POSTING_WORKER_START", logrus.Fields{
// 		"message_id": message.MessageID,
// 		"cn_id":      message.CNID,
// 		"batch_id":   message.BatchID,
// 	})

// 	e := echo.New()

// 	req := httptest.NewRequest(http.MethodPost, "/", nil)
// 	rec := httptest.NewRecorder()

// 	// set headers yang biasa dipakai SAP SDK
// 	req.Header.Set(constant.HEADER_X_COMPANY_ID, message.CompanyID)
// 	req.Header.Set(constant.HEADER_X_EMAIL, message.Email)
// 	req.Header.Set(constant.HEADER_X_ORG_DESC, message.OrgDesc)
// 	req.Header.Set(constant.HEADER_X_SALES_ORG_ID, message.SalesOrgID)

// 	echoCtx := e.NewContext(req, rec)

// 	// Call the posting usecase
// 	// Note: We pass a background context since this is async
// 	sapDocNo, err := w.UseCase.PostingWithoutReff(echoCtx, &dto.SalesReturnCNPostingRequest{
// 		CNID:        message.CNID,
// 		PostingDate: message.PostingDate,
// 		UpdatedBy:   message.UpdatedBy,
// 	})
// 	if err != nil {
// 		logstash.ErrorSync("MULTI_POSTING_WORKER_ERROR", logrus.Fields{
// 			"message_id": message.MessageID,
// 			"cn_id":      message.CNID,
// 			"batch_id":   message.BatchID,
// 			"error":      err.Error(),
// 		})

// 		errorContext := CreateContextFromMessage(message)
// 		errorContext = AddContextField(errorContext, "retry_count", message.RetryCount)
// 		errorContext = AddContextField(errorContext, "posting_date", message.PostingDate)
// 		workerType := entity.WorkerTypeRabbitMQ

// 		w.WorkerLogger.LogError(ctx, LogWorkerOption{
// 			WorkerName:  constant.WORKER_MULTI_POSTING,
// 			WorkerType:  &workerType,
// 			BatchID:     &message.BatchID,
// 			MessageID:   &message.MessageID,
// 			ReferenceID: &message.CNID,
// 			Message:     fmt.Sprintf("Failed to post CN: %s", message.CNID),
// 			Context:     errorContext,
// 			CreatedBy:   &message.UpdatedBy,
// 		}, err)

// 		return fmt.Errorf("posting failed for CN %s: %w", message.CNID, err)
// 	}

// 	// Log success
// 	logstash.Info("MULTI_POSTING_WORKER_SUCCESS", logrus.Fields{
// 		"message_id": message.MessageID,
// 		"cn_id":      message.CNID,
// 		"batch_id":   message.BatchID,
// 		"sap_doc_no": *sapDocNo,
// 	})

// 	w.Logger.Logger.WithFields(logrus.Fields{
// 		"message_id": message.MessageID,
// 		"cn_id":      message.CNID,
// 		"sap_doc_no": *sapDocNo,
// 	}).Info("multi posting completed successfully")

// 	successContext := CreateContextFromMessage(message)
// 	successContext = AddContextField(successContext, "sap_doc_no", *sapDocNo)
// 	successContext = AddContextField(successContext, "posting_date", message.PostingDate)
// 	successContext = AddContextField(successContext, "retry_count", message.RetryCount)
// 	workerType := entity.WorkerTypeRabbitMQ

// 	w.WorkerLogger.LogSuccess(ctx, LogWorkerOption{
// 		WorkerName:  constant.WORKER_MULTI_POSTING,
// 		WorkerType:  &workerType,
// 		BatchID:     &message.BatchID,
// 		MessageID:   &message.MessageID,
// 		ReferenceID: &message.CNID,
// 		Message:     fmt.Sprintf("Successfully posted CN: %s to SAP", message.CNID),
// 		Context:     successContext,
// 		CreatedBy:   &message.UpdatedBy,
// 	})

// 	return nil
// }

// func (w *MultiPostingWorker) OnError(ctx context.Context, delivery amqp.Delivery, err error) {
// 	var message dtopkg.MultiPostingWorkerMessage
// 	json.Unmarshal(delivery.Body, &message)

// 	w.Logger.Logger.WithError(err).WithFields(logrus.Fields{
// 		"message_id":   message.MessageID,
// 		"cn_id":        message.CNID,
// 		"retry_count":  message.RetryCount,
// 		"max_retry":    message.MaxRetry,
// 		"enable_retry": w.IsRetryEnabled(),
// 	}).Error("multi posting worker processing failed")

// 	// Check if we should retry
// 	if w.IsRetryEnabled() && message.RetryCount < message.MaxRetry {
// 		message.RetryCount++

// 		w.Logger.Logger.WithFields(logrus.Fields{
// 			"message_id":  message.MessageID,
// 			"cn_id":       message.CNID,
// 			"retry_count": message.RetryCount,
// 		}).Warn("Requeuing message for retry")

// 		retryContext := CreateContextFromMessage(message)
// 		retryContext = AddContextField(retryContext, "retry_count", message.RetryCount)
// 		retryContext = AddContextField(retryContext, "max_retry", message.MaxRetry)

// 		workerType := entity.WorkerTypeRabbitMQ
// 		w.WorkerLogger.LogWarning(ctx, LogWorkerOption{
// 			WorkerName:  constant.WORKER_MULTI_POSTING,
// 			WorkerType:  &workerType,
// 			BatchID:     &message.BatchID,
// 			MessageID:   &message.MessageID,
// 			ReferenceID: &message.CNID,
// 			Message:     fmt.Sprintf("Retrying CN: %s (attempt %d/%d)", message.CNID, message.RetryCount, message.MaxRetry),
// 			Context:     retryContext,
// 			CreatedBy:   &message.UpdatedBy,
// 		})

// 		// Requeue the message
// 		if err := w.requeueMessage(&message); err != nil {
// 			w.Logger.Logger.WithError(err).Error("Failed to requeue message")
// 		}
// 	}
// }

// func (w *MultiPostingWorker) requeueMessage(message *dtopkg.MultiPostingWorkerMessage) error {
// 	messageBytes, err := json.Marshal(message)
// 	if err != nil {
// 		return err
// 	}

// 	retryQueue := fmt.Sprintf(constant.WORKER_RETRY_QUEUE_DECLARE_NAME, w.QueueName)

// 	return w.QueueWrapper.Publish(
// 		retryQueue,
// 		messageBytes,
// 		5,
// 	)
// }
