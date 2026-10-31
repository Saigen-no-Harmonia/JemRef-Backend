package db

import (
	"context"
	"database/sql"
	"fmt"
	"jemref_go/internal/repository"
)

type TxManagerImpl struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManagerImpl {
	return &TxManagerImpl{db: db}
}

func (m *TxManagerImpl) WithTransaction(
	ctx context.Context,
	fn func(ctx context.Context, tx repository.DBTX) error,
) error {

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction error: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction error: %w", err)
	}

	return nil
}
