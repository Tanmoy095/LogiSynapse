package invoice

import (
	"context"
	"errors"
	"fmt"
	"testing"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/ledger"
	"github.com/google/uuid"
)

// --- MOCKS ---
// MockLedgerStore simulates fetching financial records
type MockLedgerStore struct {
	Entries []ledger.LedgerEntry
	Err     error
}

func (m *MockLedgerStore) CreateLedgerEntry(ctx context.Context, entry ledger.LedgerEntry) error {
	return nil
}

func (m *MockLedgerStore) GetEntriesForPeriod(ctx context.Context, tenantID uuid.UUID, year int, month int) ([]ledger.LedgerEntry, error) {
	return m.Entries, m.Err
}

// MockInvoiceStore simulates the database for invoices
type MockInvoiceStore struct {
	SavedInvoice    *Invoice
	ExistingInvoice *Invoice // What returns when GetInvoice is called
	DeletedID       *uuid.UUID
	ErrGet          error
	ErrDelete       error
	ErrCreate       error
	ErrUpdate       error
}

func (m *MockInvoiceStore) CreateInvoice(ctx context.Context, inv *Invoice) error {
	if m.ErrCreate != nil {
		return m.ErrCreate
	}
	m.SavedInvoice = inv
	return nil
}

func (m *MockInvoiceStore) GetInvoice(ctx context.Context, tenantID uuid.UUID, year int, month int) (*Invoice, error) {
	return m.ExistingInvoice, m.ErrGet
}

func (m *MockInvoiceStore) DeleteInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	if m.ErrDelete != nil {
		return m.ErrDelete
	}
	m.DeletedID = &invoiceID
	return nil
}

func (m *MockInvoiceStore) UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status InvoiceStatus) error {
	return m.ErrUpdate
}
func (m *MockInvoiceStore) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	return nil, nil
}
func (m *MockInvoiceStore) FinalizeInvoice(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockInvoiceStore) MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID, transactionID string) error {
	return nil
}

// Helper to create a dummy UUID for tests
var testTenantID = uuid.New()

func TestGenerateInvoiceForTenant(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name            string
		existingInvoice *Invoice // State of DB before run
		ledgerEntries   []ledger.LedgerEntry
		errGet          error // Error for GetInvoice
		errDelete       error // Error for DeleteInvoice
		errLedger       error // Error for GetEntries
		errCreate       error // Error for CreateInvoice
		expectError     error // Expected error (use for errors.Is)
		expectCreated   bool  // Should we save a new invoice?
		expectDeleted   bool  // Should we delete an old draft?
		expectInvNil    bool  // Should the returned invoice be nil?
		// Validation checks on the returned/saved invoice
		validateFunc func(t *testing.T, inv *Invoice)
	}{
		{
			name: "Happy Path: Standard Invoice Generation with Multiple Usage Types",
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     500, // $5.00
					Quantity:        10,
					UnitPrice:       50, // $0.50
					Currency:        "USD",
				},
				{
					UsageType:       billingtypes.APIRequest,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     100, // $1.00
					Quantity:        100,
					UnitPrice:       1, // $0.01
					Currency:        "USD",
				},
			},
			expectCreated: true,
			expectInvNil:  false,
			validateFunc: func(t *testing.T, inv *Invoice) {
				if inv == nil {
					t.Fatal("Expected non-nil invoice")
				}
				if inv.TotalCents != 600 {
					t.Errorf("Expected total 600 cents, got %d", inv.TotalCents)
				}
				if inv.Currency != "USD" {
					t.Errorf("Expected currency USD, got %s", inv.Currency)
				}
				if inv.Status != InvoiceDraft {
					t.Errorf("Expected status Draft, got %s", inv.Status)
				}
				if len(inv.Lines) != 2 {
					t.Errorf("Expected 2 lines, got %d", len(inv.Lines))
				} else {
					// Check first line (ShipmentCreated)
					line1 := inv.Lines[0]
					if line1.UsageType != billingtypes.ShipmentCreated {
						t.Errorf("Expected line1 UsageType ShipmentCreated, got %s", line1.UsageType)
					}
					if line1.LineTotalCents != 500 {
						t.Errorf("Expected line1 total 500, got %d", line1.LineTotalCents)
					}
					if line1.Quantity != 10 {
						t.Errorf("Expected line1 quantity 10, got %d", line1.Quantity)
					}
					if line1.UnitPriceCents != 50 {
						t.Errorf("Expected line1 unit price 50, got %d", line1.UnitPriceCents)
					}
					if line1.Description != fmt.Sprintf("%s Charges", billingtypes.ShipmentCreated) {
						t.Errorf("Expected line1 description '%s Charges', got %s", billingtypes.ShipmentCreated, line1.Description)
					}

					// Check second line (APIRequest)
					line2 := inv.Lines[1]
					if line2.UsageType != billingtypes.APIRequest {
						t.Errorf("Expected line2 UsageType APIRequest, got %s", line2.UsageType)
					}
					if line2.LineTotalCents != 100 {
						t.Errorf("Expected line2 total 100, got %d", line2.LineTotalCents)
					}
					if line2.Quantity != 100 {
						t.Errorf("Expected line2 quantity 100, got %d", line2.Quantity)
					}
					if line2.UnitPriceCents != 1 {
						t.Errorf("Expected line2 unit price 1, got %d", line2.UnitPriceCents)
					}
				}
			},
		},
		{
			name: "Regeneration: Delete Draft and Regenerate New Invoice",
			existingInvoice: &Invoice{
				InvoiceID: uuid.New(),
				Status:    InvoiceDraft, // Previous draft
			},
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     100, // $1.00
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
			},
			expectCreated: true,
			expectDeleted: true, // MUST delete the old draft
			expectInvNil:  false,
			validateFunc: func(t *testing.T, inv *Invoice) {
				if inv.TotalCents != 100 {
					t.Errorf("Expected total 100 cents, got %d", inv.TotalCents)
				}
				if len(inv.Lines) != 1 {
					t.Errorf("Expected 1 line, got %d", len(inv.Lines))
				}
			},
		},
		{
			name: "Security Guard: Cannot Regenerate Finalized Invoice",
			existingInvoice: &Invoice{
				InvoiceID: uuid.New(),
				Status:    InvoiceFinalized, // Immutable!
			},
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
			},
			expectError:   ErrInvoiceAlreadyFinalized,
			expectCreated: false,
			expectDeleted: false,
			expectInvNil:  true,
		},
		{
			name: "Currency Safety: Fail on Mixed Currencies",
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "EUR",
				},
			},
			expectError:   ErrCurrencyMismatch,
			expectCreated: false,
			expectDeleted: false,
			expectInvNil:  true,
		},
		{
			name: "Complex Logic: Variable Pricing (Tiered) for Same Usage Type",
			// Scenario: 10 units @ $1.00, then 10 units @ $0.80.
			// Result: Total $18.00, Quantity 20, Unit Price 0 (Variable)
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     1000,
					Quantity:        10,
					UnitPrice:       100,
					Currency:        "USD",
				},
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     800,
					Quantity:        10,
					UnitPrice:       80, // Different price!
					Currency:        "USD",
				},
			},
			expectCreated: true,
			expectInvNil:  false,
			validateFunc: func(t *testing.T, inv *Invoice) {
				if len(inv.Lines) != 1 {
					t.Fatalf("Expected 1 aggregated line, got %d", len(inv.Lines))
				}
				line := inv.Lines[0]
				if line.LineTotalCents != 1800 {
					t.Errorf("Expected line total 1800, got %d", line.LineTotalCents)
				}
				if line.Quantity != 20 {
					t.Errorf("Expected quantity 20, got %d", line.Quantity)
				}
				// CRITICAL CHECK: Unit Price should be 0 because it varied
				if line.UnitPriceCents != 0 {
					t.Errorf("Expected variable unit price to be 0, got %d", line.UnitPriceCents)
				}
				if inv.TotalCents != 1800 {
					t.Errorf("Expected invoice total 1800, got %d", inv.TotalCents)
				}
			},
		},
		{
			name:          "Empty Ledger: No Invoice Generated",
			ledgerEntries: []ledger.LedgerEntry{}, // Empty
			expectCreated: false,
			expectDeleted: false,
			expectInvNil:  true,
		},
		{
			name: "With Credits: Net Total Calculation for Same Usage Type",
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     500,
					Quantity:        5,
					UnitPrice:       100,
					Currency:        "USD",
				},
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeCredit,
					AmountCents:     200,
					Quantity:        2,
					UnitPrice:       100,
					Currency:        "USD",
				},
			},
			expectCreated: true,
			expectInvNil:  false,
			validateFunc: func(t *testing.T, inv *Invoice) {
				if inv.TotalCents != 300 {
					t.Errorf("Expected total 300 cents, got %d", inv.TotalCents)
				}
				if len(inv.Lines) != 1 {
					t.Errorf("Expected 1 line, got %d", len(inv.Lines))
				} else {
					line := inv.Lines[0]
					if line.LineTotalCents != 300 {
						t.Errorf("Expected line total 300, got %d", line.LineTotalCents)
					}
					if line.Quantity != 7 {
						t.Errorf("Expected quantity 7, got %d", line.Quantity)
					}
					if line.UnitPriceCents != 100 {
						t.Errorf("Expected unit price 100, got %d", line.UnitPriceCents)
					}
				}
			},
		},
		{
			name: "Zero Net Line: Debit and Credit Cancel Out for Same Type - No Line Added",
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeCredit,
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
			},
			expectCreated: true,
			expectInvNil:  false,
			validateFunc: func(t *testing.T, inv *Invoice) {
				if inv.TotalCents != 0 {
					t.Errorf("Expected total 0 cents, got %d", inv.TotalCents)
				}
				if len(inv.Lines) != 0 {
					t.Errorf("Expected 0 lines (zero net), got %d", len(inv.Lines))
				}
			},
		},
		{
			name: "Zero Net Invoice: Debits and Credits on Different Types",
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
				{
					UsageType:       billingtypes.APIRequest,
					TransactionType: billingtypes.TransactionTypeCredit,
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
			},
			expectCreated: true,
			expectInvNil:  false,
			validateFunc: func(t *testing.T, inv *Invoice) {
				if inv.TotalCents != 0 {
					t.Errorf("Expected total 0 cents, got %d", inv.TotalCents)
				}
				if len(inv.Lines) != 2 {
					t.Errorf("Expected 2 lines, got %d", len(inv.Lines))
				} else {
					// Check debit line positive
					var debitLine, creditLine InvoiceLine
					if inv.Lines[0].LineTotalCents > 0 {
						debitLine = inv.Lines[0]
						creditLine = inv.Lines[1]
					} else {
						debitLine = inv.Lines[1]
						creditLine = inv.Lines[0]
					}
					if debitLine.LineTotalCents != 100 {
						t.Errorf("Expected debit line total 100, got %d", debitLine.LineTotalCents)
					}
					if creditLine.LineTotalCents != -100 {
						t.Errorf("Expected credit line total -100, got %d", creditLine.LineTotalCents)
					}
				}
			},
		},
		{
			name: "Unknown Transaction Type: Skipped in Aggregation",
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: "Unknown", // Invalid
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
				{
					UsageType:       billingtypes.APIRequest,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     200,
					Quantity:        2,
					UnitPrice:       100,
					Currency:        "USD",
				},
			},
			expectCreated: true,
			expectInvNil:  false,
			validateFunc: func(t *testing.T, inv *Invoice) {
				if inv.TotalCents != 200 {
					t.Errorf("Expected total 200 cents (unknown skipped), got %d", inv.TotalCents)
				}
				if len(inv.Lines) != 1 {
					t.Errorf("Expected 1 line, got %d", len(inv.Lines))
				}
			},
		},
		{
			name:          "Error: Failed to Check Existing Invoice",
			errGet:        errors.New("database connection failed"),
			expectError:   errors.New("failed to check existing invoice: database connection failed"), // Exact wrapped error
			expectCreated: false,
			expectDeleted: false,
			expectInvNil:  true,
		},
		{
			name: "Error: Failed to Delete Old Draft",
			existingInvoice: &Invoice{
				InvoiceID: uuid.New(),
				Status:    InvoiceDraft,
			},
			errDelete:     errors.New("delete failed"),
			expectError:   errors.New("failed to delete old draft: delete failed"),
			expectCreated: false,
			expectDeleted: false,
			expectInvNil:  true,
		},
		{
			name:          "Error: Failed to Fetch Ledger Entries",
			errLedger:     errors.New("ledger query error"),
			expectError:   errors.New("failed to fetch ledger entries: ledger query error"),
			expectCreated: false,
			expectDeleted: false,
			expectInvNil:  true,
		},
		{
			name: "Error: Failed to Create New Invoice",
			ledgerEntries: []ledger.LedgerEntry{
				{
					UsageType:       billingtypes.ShipmentCreated,
					TransactionType: billingtypes.TransactionTypeDebit,
					AmountCents:     100,
					Quantity:        1,
					UnitPrice:       100,
					Currency:        "USD",
				},
			},
			errCreate:     errors.New("insert failed"),
			expectError:   errors.New("failed to save invoice: insert failed"),
			expectCreated: false,
			expectDeleted: false,
			expectInvNil:  true,
		},
	}

	// EXECUTE THE TESTS
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Setup Mocks
			mockLedger := &MockLedgerStore{Entries: tt.ledgerEntries, Err: tt.errLedger}
			mockInvoice := &MockInvoiceStore{
				ExistingInvoice: tt.existingInvoice,
				ErrGet:          tt.errGet,
				ErrDelete:       tt.errDelete,
				ErrCreate:       tt.errCreate,
			}
			generator := NewInvoiceGenerator(mockLedger, mockInvoice)

			// 2. Run Logic
			inv, err := generator.GenerateInvoiceForTenant(context.Background(), testTenantID, 2024, 6)

			// 3. Verify Error State
			if tt.expectError != nil {
				if err == nil {
					t.Errorf("Expected error containing '%v', got nil", tt.expectError)
				} else if !errors.Is(err, tt.expectError) && err.Error() != tt.expectError.Error() {
					t.Errorf("Expected error '%v', got '%v'", tt.expectError, err)
				}
				// For wrapped errors, we allow string match as fallback, but primarily use Is
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// 4. Verify Returned Invoice
			if tt.expectInvNil {
				if inv != nil {
					t.Error("Expected returned invoice to be nil, but got non-nil")
				}
			} else {
				if inv == nil {
					t.Error("Expected returned invoice to be non-nil, but got nil")
				}
			}

			// 5. Verify Creation Logic
			if tt.expectCreated {
				if mockInvoice.SavedInvoice == nil {
					t.Error("Expected invoice to be saved, but it wasn't")
				} else {
					// Run custom validation if provided
					if tt.validateFunc != nil {
						tt.validateFunc(t, inv) // Validate on returned inv
					}
				}
			} else {
				if mockInvoice.SavedInvoice != nil {
					t.Error("Expected NO invoice to be saved, but one was created")
				}
			}

			// 6. Verify Deletion Logic
			if tt.expectDeleted {
				if mockInvoice.DeletedID == nil {
					t.Error("Expected previous draft to be deleted, but it wasn't")
				} else if *mockInvoice.DeletedID != tt.existingInvoice.InvoiceID {
					t.Errorf("Deleted wrong invoice ID: expected %v, got %v", tt.existingInvoice.InvoiceID, *mockInvoice.DeletedID)
				}
			} else {
				if mockInvoice.DeletedID != nil {
					t.Error("Expected NO deletion, but DeleteInvoice was called")
				}
			}
		})
	}
}
