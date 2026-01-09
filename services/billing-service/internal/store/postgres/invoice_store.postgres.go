//services/billing-service/internal/store/postgres/invoice_store.postgres.go

package PostgresStore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/invoice"
	"github.com/google/uuid"
)

type PostgresInvoiceStore struct {
	db *sql.DB
}

func NewPostgresInvoiceStore(db *sql.DB) *PostgresInvoiceStore {
	return &PostgresInvoiceStore{db: db}
}

// Implement InvoiceStore methods here
// CreateInvoice inserts a new invoice into the store.
// GetInvoice retrieves an invoice by tenant ID, year, and month.
// UpdateInvoice updates an existing invoice in the store.
func (store *PostgresInvoiceStore) CreateInvoice(ctx context.Context, inv *invoice.Invoice) error {
	//Begin a transaction.
	tx, err := store.db.BeginTx(ctx, nil) //store.db.BeginTx starts a new transaction with the provided context. means any operation within this transaction can be cancelled or timed out based on the context.
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	//defer a rollback in case anything fails.
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 2. Insert Header (The Invoice itself)
	headerQuery := `
		INSERT INTO invoices (invoice_id, tenant_id, billing_year, billing_month, total_amount_cents,currency,status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.ExecContext(ctx, headerQuery,
		inv.InvoiceID,
		inv.TenantID,
		inv.Year,
		inv.Month,
		inv.TotalCents,
		inv.Currency,
		inv.Status,
		inv.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert invoice header: %w", err)
	}
	//Insert Lines
	lineQuery := `
		INSERT INTO invoice_lines (id, invoice_id, usage_type, quantity, unit_price_cents, line_total_cents, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	// We use the prepared statement for efficiency in loops
	stmt, err := tx.PrepareContext(ctx, lineQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare line statement: %w", err)
	}
	defer stmt.Close()

	for _, line := range inv.Lines {
		_, err = stmt.ExecContext(ctx,
			line.ID,
			inv.InvoiceID,
			line.UsageType,
			line.Quantity,
			line.UnitPriceCents,
			line.LineTotalCents,
			line.Description,
		)
		if err != nil {
			return fmt.Errorf("failed to insert invoice line: %w", err)
		}
	}

	// Commit the transaction if all operations succeeded
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// GetInvoice fetches the header AND the lines, reassembling them into the struct
func (store *PostgresInvoiceStore) GetInvoice(ctx context.Context, tenantID uuid.UUID, year int, month int) (*invoice.Invoice, error) {
	// Fetch Invoice Header
	headerQuery := `
		SELECT invoice_id, tenant_id, billing_year, billing_month, total_amount_cents, currency, status, created_at,finalized_at, paid_at
		FROM invoices
		WHERE tenant_id = $1 AND billing_year = $2 AND billing_month = $3`

	var inv invoice.Invoice
	// We use sql.NullTime for nullable timestamps
	var finalizedAt, paidAt sql.NullTime
	err := store.db.QueryRowContext(ctx, headerQuery, tenantID, year, month).Scan(
		&inv.InvoiceID,
		&inv.TenantID,
		&inv.Year,
		&inv.Month,
		&inv.TotalCents,
		&inv.Currency,
		&inv.Status,
		&inv.CreatedAt,
		&finalizedAt,
		&paidAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error here, just nil
		}
		return nil, fmt.Errorf("failed to fetch invoice header: %w", err)
	}

	// B. Fetch Lines
	linesQuery := `
		SELECT id, usage_type, quantity, unit_price_cents, line_total_cents, description
		FROM invoice_lines
		WHERE invoice_id = $1
	`
	rows, err := store.db.QueryContext(ctx, linesQuery, inv.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice lines: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var line invoice.InvoiceLine
		var uType string // Scan as string, cast to Enum

		if err := rows.Scan(
			&line.ID,
			&uType,
			&line.Quantity,
			&line.UnitPriceCents,
			&line.LineTotalCents,
			&line.Description,
		); err != nil {
			return nil, err
		}
		line.UsageType = billingtypes.UsageType(uType)
		inv.Lines = append(inv.Lines, line)
	}

	return &inv, nil
}

// DeleteInvoice allows us to clear a DRAFT so we can regenerate it cleanly
func (s *PostgresInvoiceStore) DeleteInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	// ON DELETE CASCADE in SQL should handle the lines, but safe to be explicit or rely on FK
	query := `DELETE FROM invoices WHERE invoice_id = $1`
	_, err := s.db.ExecContext(ctx, query, invoiceID)
	return err
}

// UpdateStatus changes the status (e.g., DRAFT -> FINALIZED)
func (s *PostgresInvoiceStore) UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status invoice.InvoiceStatus) error {
	query := `
		UPDATE invoices 
		SET status = $1, finalized_at = CASE WHEN $1 = 'FINALIZED' THEN NOW() ELSE finalized_at END
		WHERE invoice_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, status, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}
	return nil
}

func (s *PostgresInvoiceStore) GetInvoiceByID(ctx context.Context, invoiceID uuid.UUID) (*invoice.Invoice, error) {
	//Fetch Invoice by its UUID
	query := ` SELECT invoice_id, tenant_id, billing_year, billing_month, total_amount_cents, currency, status, created_at,finalized_at, paid_at
	FROM invoices
	WHERE invoice_id = $1
	`
	var inv invoice.Invoice
	// We use sql.NullTime for nullable timestamps
	var finalizedAt, paidAt sql.NullTime
	err := s.db.QueryRowContext(ctx, query, invoiceID).Scan(
		&inv.InvoiceID,
		&inv.TenantID,
		&inv.Year,
		&inv.Month,
		&inv.TotalCents,
		&inv.Currency,
		&inv.Status,
		&inv.CreatedAt,
		&finalizedAt,
		&paidAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error here, just nil
		}
		return nil, fmt.Errorf("failed to fetch invoice by ID: %w", err)
	}
	// 2. Fetch Lines (Reusing existing logic logic is fine, or simple query)
	// For finalization validation, we specifically need to know if lines exist.
	lineQuery := `SELECT count(*) FROM invoice_lines WHERE invoice_id = $1`
	var lineCount int
	if err := s.db.QueryRowContext(ctx, lineQuery, inv.InvoiceID).Scan(&lineCount); err != nil {
		return nil, fmt.Errorf("db count lines failed: %w", err)
	}
	if lineCount == 0 {
		return nil, fmt.Errorf("invoice %s has no lines", invoiceID)
	}
	//We verify line count in the Service layer, but fetching the actual lines is optional
	// (If you need full lines, copy the fetching logic from GetInvoice).

	return &inv, nil

}

func (s *PostgresInvoiceStore) FinalizeInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	// we only allow finalization if the invoice is currently in DRAFT status
	// This prevents "Double Finalization" at the DB level.

	query := `
		UPDATE invoices 
		SET status = 'FINALIZED', finalized_at = NOW()
		WHERE invoice_id = $1 AND status = 'DRAFT'
	`
	res, err := s.db.ExecContext(ctx, query, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to finalize invoice: %w", err)
	}
	rows, err := res.RowsAffected() // this ensure that only one row was updated
	if err != nil {
		return fmt.Errorf("failed to check rows affected during finalization: %w", err)
	}
	if rows == 0 {
		// If 0 rows updated, it means either:
		// 1. Invoice doesn't exist
		// 2. Invoice exists but is NOT in DRAFT state
		// We check existence to return the right error.
		existsQuery := `SELECT status FROM invoices WHERE invoice_id = $1`
		var status string
		if err := s.db.QueryRowContext(ctx, existsQuery, invoiceID).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return invoice.ErrInvoiceNotFound
			}
			return err
		}

		// If we are here, it exists but wasn't DRAFT
		return invoice.ErrInvoiceNotDraft
	}
	return nil
}

// -------------------------------!!After Payment Phase !!-----------------------------------
// MarkInvoicePaid marks an invoice as PAID, only if it is currently FINALIZED
func (s *PostgresInvoiceStore) MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID, transactionID string) error {
	// we only allow marking as PAID if the invoice is currently in FINALIZED status
	query := `
		UPDATE invoices 
		SET status = 'PAID', 
		payment_intent_id = $2,
		paid_at = NOW(),
		updated_at = NOW()
		WHERE invoice_id = $1 AND status = 'FINALIZED'
	`
	res, err := s.db.ExecContext(ctx, query, invoiceID, transactionID)
	if err != nil {
		return fmt.Errorf("db:failed to mark invoice as PAID: %w", err)
	}
	rows, err := res.RowsAffected() // this ensure that only one row was updated
	if err != nil {
		return fmt.Errorf("db: failed to check rows affected: %w", err)
	}

	// 0 Rows means the invoice either doesn't exist OR it wasn't in FINALIZED state.
	// This protects us from paying a DRAFT or an already PAID invoice.
	if rows == 0 {
		return fmt.Errorf("invoice not found or not in FINALIZED state")
	}

	return nil
}
