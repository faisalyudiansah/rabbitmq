package entity

import "time"

type Job struct {
	ID        int64      `db:"id"`
	Type      string     `db:"type"`
	Payload   string     `db:"payload"`
	Status    string     `db:"status"`
	Retry     int        `db:"retry"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
