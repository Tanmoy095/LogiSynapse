// services/authentication-service/internal/infra/postgres/tx_manager.postgres.go
package postgres

import (
	"context"
	"database/sql"
)

type TxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

type txKey struct{}

func (tm *TxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return err
	}

	// Ensure rollback on panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Inject tx into context
	ctx = context.WithValue(ctx, txKey{}, tx)

	err = fn(ctx)
	if err != nil {
		return err
	}

	return tx.Commit()
}
