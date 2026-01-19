package dto

import "time"

// Base Worker Message
type WorkerMessage struct {
	MessageID  string    `json:"message_id"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
	RetryCount int       `json:"retry_count"`
	MaxRetry   int       `json:"max_retry"`
}

type HeaderHTTPRequest struct {
	CompanyID  string `json:"X-Company-Id,omitempty"`
	Email      string `json:"X-Email,omitempty"`
	OrgDesc    string `json:"X-Org-Desc,omitempty"`
	SalesOrgID string `json:"X-Sales-Org-Id,omitempty"`
}

// Multi Posting Worker Message
type MultiPostingWorkerMessage struct {
	WorkerMessage
	CNID        string    `json:"cn_id"`
	PostingDate time.Time `json:"posting_date"`
	UpdatedBy   string    `json:"updated_by"`

	BatchID      string `json:"batch_id,omitempty"`
	TotalInBatch int    `json:"total_in_batch,omitempty"`
	IndexInBatch int    `json:"index_in_batch,omitempty"`

	HeaderHTTPRequest
}

// Email Worker Message (example for future)
type EmailWorkerMessage struct {
	WorkerMessage
	To       []string               `json:"to"`
	Subject  string                 `json:"subject"`
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

// Notification Worker Message (example for future)
type NotificationWorkerMessage struct {
	WorkerMessage
	UserID  string                 `json:"user_id"`
	Title   string                 `json:"title"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
