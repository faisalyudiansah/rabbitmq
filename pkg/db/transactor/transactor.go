package transactor

import (
	"context"
	"database/sql"
)

type TransactorInterface interface {
	Atomic(ctx context.Context, fn func(context.Context) error) error
}

type transactorImpl struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) *transactorImpl {
	return &transactorImpl{
		db: db,
	}
}

func (t *transactorImpl) Atomic(ctx context.Context, fn func(context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = fn(injectTx(ctx, tx))
	if err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			return err
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

type TxKey struct{}

func injectTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, TxKey{}, tx)
}

func ExtractTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(TxKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}
