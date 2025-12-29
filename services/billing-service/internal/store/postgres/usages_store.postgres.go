package postgres

import (
	"context"
	"database/sql"
	"fmt"

	storepkg "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/store"
)

type PostgresUsageStore struct { // db is the database connection (placeholder, implement as needed)
	db *sql.DB
}

func NewPostgresUsageStore(db *sql.DB) *PostgresUsageStore {
	return &PostgresUsageStore{db: db}
}
func (store *PostgresUsageStore) Flush(ctx context.Context, batch storepkg.FlushBatch) error {
	// start Transaction with context support (support for cancellation and timeouts)
	//Why use Begin TX ? if anything fails later you can rollback and undo all changes .If everything success you commit and make he changes permanent
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	//Defer rollback on error /panic .commit only on success
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	// Idempotency check : Insert BatchID; on conflict do nothing and return success (no-op)
	res, err := tx.ExecContext(ctx, `
		INSERT INTO flush_history (batch_id)
		VALUES ($1) 
		ON CONFLICT (batch_id) DO NOTHING
	`, batch.BatchID)
	if err != nil {
		return fmt.Errorf("failed to insert flush history Idempotency Insert: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for flush history insert: %w", err)
	}
	if rowsAffected == 0 {
		// Batch ID already exists, indicating this batch has been processed
		return tx.Commit() // Commit to finalize the transaction
	}
	//Prepare upsert statement for efficiency

	//upsert means insert or update if exists
	stmt, err := tx.PrepareContext(ctx, `
	INSERT INTO usage_aggregates (tenant_id, usage_type, billing_year, billing_month, total_quantity)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (tenant_id, usage_type, billing_year, billing_month)
	DO UPDATE SET total_quantity = usage_aggregates.total_quantity + EXCLUDED.total_quantity,
	last_updated = NOW()

	`)
	if err != nil {
		return fmt.Errorf("failed to prepare upsert statement: %w", err)
	}
	defer stmt.Close() //Ensure statement is closed after use

	// Iterate over usage records and execute upsert for each
	for _, record := range batch.Records {
		_, err := stmt.ExecContext(ctx,
			record.TenantID,
			record.UsageType,
			record.BillingPeriod.Year,
			record.BillingPeriod.Month,
			record.TotalQuantity,
		)
		if err != nil {
			return fmt.Errorf("failed to execute upsert for tenant %s: %w", record.TenantID, err)
		}
	}
	//commit if all successful
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil

}
