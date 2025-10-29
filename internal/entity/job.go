package entity

import "time"

type Job struct {
	ID         int64      `db:"id" gorm:"primaryKey"`
	Type       string     `db:"type"`
	Payload    string     `db:"payload"`
	Status     string     `db:"status"`
	RetryCount int        `db:"retry_count"`
	MaxRetry   int        `db:"max_retry"`
	LastError  *string    `db:"last_error"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
}
