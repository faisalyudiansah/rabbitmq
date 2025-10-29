package usecase

import (
	"background-job-service/internal/entity"
	"background-job-service/internal/repository"
	"background-job-service/pkg/db/transactor"
	"context"
	"encoding/json"
)

type JobUseCaseInterface interface {
	CreateJob(ctx context.Context, jobType string, payload any) (*int64, error)
	UpdateJobStatus(ctx context.Context, id int64, status string) error
	IncrementRetry(ctx context.Context, id int64) error
	GetJobByID(ctx context.Context, id int64) (*entity.Job, error)
}

type jobUseCase struct {
	transactor transactor.TransactorInterface
	repo       repository.JobRepositoryInterface
}

func NewJobUseCase(trx transactor.TransactorInterface, repo repository.JobRepositoryInterface) *jobUseCase {
	return &jobUseCase{
		transactor: trx,
		repo:       repo,
	}
}

func (u *jobUseCase) CreateJob(ctx context.Context, jobType string, payload any) (*int64, error) {
	var jobID int64
	err := u.transactor.Atomic(ctx, func(cForTx context.Context) error {
		bytes, _ := json.Marshal(payload)
		job := &entity.Job{
			Type:    jobType,
			Payload: string(bytes),
			Status:  "pending",
		}
		id, err := u.repo.Create(cForTx, job)
		jobID = id
		return err
	})
	if err != nil {
		return nil, err
	}
	return &jobID, nil
}

func (u *jobUseCase) UpdateJobStatus(ctx context.Context, id int64, status string) error {
	return u.repo.UpdateStatus(ctx, id, status)
}

func (u *jobUseCase) IncrementRetry(ctx context.Context, id int64) error {
	return u.repo.IncrementRetry(ctx, id)
}

func (u *jobUseCase) GetJobByID(ctx context.Context, id int64) (*entity.Job, error) {
	return u.repo.GetByID(ctx, id)
}
