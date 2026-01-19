package repositoryworker

import (
	entity "background-job-service/internal/entity/worker"
	"context"
	"database/sql"
	"encoding/json"
	"strings"
)

type LogWorkerRepositoryInterface interface {
	Create(ctx context.Context, e *entity.LogWorker) error
	Update(ctx context.Context, e *entity.LogWorker) error
	List(ctx context.Context, f *entity.LogWorkerFilter) ([]entity.LogWorker, error)
}

type LogWorkerRepository struct {
	db *sql.DB
}

type LogWorkerRepositoryOption struct {
	db *sql.DB
}

func NewLogWorkerRepository(opt LogWorkerRepositoryOption) *LogWorkerRepository {
	return &LogWorkerRepository{
		db: opt.db,
	}
}

func (r *LogWorkerRepository) Create(ctx context.Context, e *entity.LogWorker) error {
	query := `
		INSERT INTO service_sales_return.log_workers (
			worker_name,
			worker_type,
			batch_id,
			message_id,
			reference_id,
			level,
			message,
			error,
			context,
			created_at,
			created_by
		) VALUES (?,?,?,?,?,?,?,?,?,NOW(),?)
	`

	var contextJSON []byte
	if e.Context != nil {
		var err error
		contextJSON, err = json.Marshal(e.Context)
		if err != nil {
			return err
		}
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		e.WorkerName,
		e.WorkerType,
		e.BatchID,
		e.MessageID,
		e.ReferenceID,
		e.Level,
		e.Message,
		e.Error,
		contextJSON,
		e.CreatedBy,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		e.ID = id
	}

	return nil
}

func (r *LogWorkerRepository) Update(ctx context.Context, e *entity.LogWorker) error {
	query := `
		UPDATE service_sales_return.log_workers
		SET
			worker_name = ?,
			worker_type = ?,
			batch_id = ?,
			message_id = ?,
			reference_id = ?,
			level = ?,
			message = ?,
			error = ?,
			context = ?
		WHERE id = ?
	`

	var contextJSON []byte
	if e.Context != nil {
		var err error
		contextJSON, err = json.Marshal(e.Context)
		if err != nil {
			return err
		}
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		e.WorkerName,
		e.WorkerType,
		e.BatchID,
		e.MessageID,
		e.ReferenceID,
		e.Level,
		e.Message,
		e.Error,
		contextJSON,
		e.ID,
	)

	return err
}

func (r *LogWorkerRepository) List(ctx context.Context, f *entity.LogWorkerFilter) ([]entity.LogWorker, error) {

	baseQuery := `
		SELECT
			id,
			worker_name,
			worker_type,
			batch_id,
			message_id,
			reference_id,
			level,
			message,
			error,
			context,
			created_at,
			created_by
		FROM service_sales_return.log_workers
	`

	var (
		conditions []string
		args       []interface{}
	)

	if f != nil {
		if f.WorkerName != nil {
			conditions = append(conditions, "worker_name = ?")
			args = append(args, *f.WorkerName)
		}
		if f.BatchID != nil {
			conditions = append(conditions, "batch_id = ?")
			args = append(args, *f.BatchID)
		}
		if f.ReferenceID != nil {
			conditions = append(conditions, "reference_id = ?")
			args = append(args, *f.ReferenceID)
		}
		if f.MessageID != nil {
			conditions = append(conditions, "message_id = ?")
			args = append(args, *f.MessageID)
		}
		if f.Level != nil {
			conditions = append(conditions, "level = ?")
			args = append(args, *f.Level)
		}
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []entity.LogWorker

	for rows.Next() {
		var (
			log         entity.LogWorker
			contextJSON []byte
		)

		err := rows.Scan(
			&log.ID,
			&log.WorkerName,
			&log.WorkerType,
			&log.BatchID,
			&log.MessageID,
			&log.ReferenceID,
			&log.Level,
			&log.Message,
			&log.Error,
			&contextJSON,
			&log.CreatedAt,
			&log.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		if contextJSON != nil {
			_ = json.Unmarshal(contextJSON, &log.Context)
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}
