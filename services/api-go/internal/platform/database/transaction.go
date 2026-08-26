package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)
// TxManager defines an interface for managing database transactions.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

type txManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) TxManager {
	return &txManager{pool: pool}
}

func (m *txManager) Do(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
