// services/billing-service/internal/store/postgres/ledger_store.postgres.go

package PostgresStore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/ledger"
	"github.com/google/uuid"
)

type PostgresLedgerStore struct {
	db *sql.DB
}

func NewPostgresLedgerStore(db *sql.DB) *PostgresLedgerStore {
	return &PostgresLedgerStore{
		db: db,
	}
}

func (store *PostgresLedgerStore) CreateLedgerEntry(ctx context.Context, entry ledger.LedgerEntry) error {
	query := `
  INSERT INTO billing_ledger 
  (tenant_id,  transaction_type,reference_id, amount_cents, usage_type, currency, description,quantity,unit_price_cents, created_at)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
  ON CONFLICT (tenant_id, reference_id) DO NOTHING;
  
  `
	// We do NOT pass entry.Timestamp because the SQL uses NOW()
	res, err := store.db.ExecContext(ctx, query,
		entry.TenantID,
		entry.TransactionType,
		entry.EntryID,
		entry.AmountCents,
		entry.UsageType,
		entry.Currency,
		entry.Description,
		entry.Quantity,
		entry.UnitPrice,
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
func (store *PostgresLedgerStore) GetLedgerEntriesForPeriod(ctx context.Context, tenantID uuid.UUID, year int, month int) ([]ledger.LedgerEntry, error) {
	// Implementation goes here
	query := `
	SELECT tenant_id, transaction_type, reference_id, amount_cents, usage_type, currency, description,quantity,unit_price_cents, created_at
	FROM billing_ledger
	WHERE tenant_id = $1
	AND EXTRACT(YEAR FROM created_at) = $2
	AND EXTRACT(MONTH FROM created_at) = $3;
	`
	rows, err := store.db.QueryContext(ctx, query, tenantID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to query ledger entries: %w", err)
	}
	defer rows.Close()

	var entries []ledger.LedgerEntry
	for rows.Next() {
		var entry ledger.LedgerEntry
		var createdAt sql.NullTime
		err := rows.Scan(
			&entry.TenantID,
			&entry.TransactionType,
			&entry.EntryID,
			&entry.AmountCents,
			&entry.Currency,
			&entry.Description,
			&entry.Quantity,
			&entry.UnitPrice,
			&createdAt,
			&entry.UsageType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over ledger entries: %w", err)
	}

	return entries, nil
}
