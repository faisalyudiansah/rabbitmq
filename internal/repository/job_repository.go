package repository

import (
	"background-job-service/internal/entity"
	"background-job-service/pkg/db/transactor"
	"context"
	"database/sql"
	"time"
)

type JobRepositoryInterface interface {
	Create(ctx context.Context, job *entity.Job) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	IncrementRetry(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entity.Job, error)
}

type jobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *jobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) Create(ctx context.Context, job *entity.Job) (int64, error) {
	var id int64
	var err error

	tx := transactor.ExtractTx(ctx)
	query := `INSERT INTO jobs (type, payload, status) VALUES ($1, $2, $3) RETURNING id`

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, job.Type, job.Payload, job.Status).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx, query, job.Type, job.Payload, job.Status).Scan(&id)
	}
	return id, err
}

func (r *jobRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	var err error

	tx := transactor.ExtractTx(ctx)
	query := `UPDATE jobs SET status = $1, updated_at = $2 WHERE id = $3`

	if tx != nil {
		_, err = tx.ExecContext(ctx, query, status, time.Now(), id)
	} else {
		_, err = r.db.ExecContext(ctx, query, status, time.Now(), id)
	}
	return err
}

func (r *jobRepository) IncrementRetry(ctx context.Context, id int64) error {
	var err error

	tx := transactor.ExtractTx(ctx)
	query := `UPDATE jobs SET retry = retry + 1, updated_at = $1 WHERE id = $2`

	if tx != nil {
		_, err = tx.ExecContext(ctx, query, time.Now(), id)
	} else {
		_, err = r.db.ExecContext(ctx, query, time.Now(), id)
	}
	return err
}

func (r *jobRepository) GetByID(ctx context.Context, id int64) (*entity.Job, error) {
	var job entity.Job
	var err error

	tx := transactor.ExtractTx(ctx)
	query := `SELECT * FROM jobs WHERE id = $1`

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, id).Scan(&job)
	} else {
		err = r.db.QueryRowContext(ctx, query, id).Scan(&job)
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}
