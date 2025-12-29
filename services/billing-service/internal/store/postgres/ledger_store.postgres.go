// services/billing-service/internal/store/postgres/ledger_store.postgres.go

package Postgres_Store

import (
	"context"
	"database/sql"
	"fmt"

	store "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/store"
)

type PostgresLedgerStore struct {
	db *sql.DB
}

func NewPostgresLedgerStore(db *sql.DB) *PostgresLedgerStore {
	return &PostgresLedgerStore{
		db: db,
	}
}

func (store *PostgresLedgerStore) CreateLedgerEntry(ctx context.Context, entry store.LedgerEntry) error {
	query := `
  INSERT INTO billing_ledger 
  (tenant_id,  transaction_type,reference_id, amount_cents, currency, description, created_at)
  VALUES ($1, $2, $3, $4, $5, $6, NOW())
  ON CONFLICT (tenant_id, reference_id) DO NOTHING;
  
  `
	// We do NOT pass entry.Timestamp because the SQL uses NOW()
	res, err := store.db.ExecContext(ctx, query,
		entry.TenantID,
		entry.TransactionType,
		entry.EntryID,
		entry.AmountCents,
		entry.Currency,
		entry.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ledger entry: %w", err)
	}

	// Optional: Check rows affected if you want to know if it was a new insert or a skip
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// This is NOT an error. It means Idempotency worked.
		// We return nil so the system knows "this is handled".
		fmt.Println("Ledger entry already exists, skipping duplicate insert.")
		return nil
	}

	return nil
}
