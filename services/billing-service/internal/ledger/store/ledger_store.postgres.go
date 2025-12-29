// services/billing-service/internal/ledger/store/ledger_store.postgres.go

package ledger_store

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresLedgerStore struct {
	db *sql.DB
}

func NewPostgresLedgerStore(db *sql.DB) *PostgresLedgerStore {
	return &PostgresLedgerStore{
		db: db,
	}
}

func (store *PostgresLedgerStore) CreateLedgerEntry(ctx context.Context, entry LedgerEntry) error {
	query := `
  INSERT INTO billing_ledger 
  (tenant_id,  transaction_type,reference_id, amount_cents, currency, description, created_at)
  VALUES ($1, $2, $3, $4, $5, $6, NOW())
  ON CONFLICT (tenant_id, reference_id) DO NOTHING;
  
  `
	res, err := store.db.ExecContext(ctx, query,
		entry.TenantID,
		entry.TransactionType,
		entry.EntryID,
		entry.AmountCents,
		entry.Currency,
		entry.Description,
		entry.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ledger entry: %w", err)
	}

	// Optional: Check rows affected if you want to know if it was a new insert or a skip
	rows, _ := res.RowsAffected()
	if rows == 0 {
		fmt.Printf("ℹ️ Ledger entry skipped (already billed): %s\n", entry.EntryID)
	}

	return nil
}
