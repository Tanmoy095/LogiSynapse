//services/billing-service/internal/invoice/invoice_finalizer_test.go

package invoice

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// --- MOCKS ---
// We reuse the MockInvoiceStore logic but tailor it for Finalization tests.

type MockFinalizerStore struct {
	// State to simulate DB content
	Invoice     *Invoice
	FetchErr    error
	FinalizeErr error
}

func (m *MockFinalizerStore) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	if m.FetchErr != nil {
		return nil, m.FetchErr
	}
	if m.Invoice == nil {
		return nil, ErrInvoiceNotFound
	}
	return m.Invoice, nil
}

func (m *MockFinalizerStore) FinalizeInvoice(ctx context.Context, id uuid.UUID) error {
	if m.FinalizeErr != nil {
		return m.FinalizeErr
	}
	// Simulate the SQL constraint: "WHERE status = 'DRAFT'"
	// If the mock invoice in memory isn't draft, the DB update would affect 0 rows
	// and return ErrInvoiceNotDraft (as per our postgres implementation).
	if m.Invoice.Status != InvoiceDraft {
		return ErrInvoiceNotDraft
	}

	// Success: Mutate the mock to match what DB would do
	m.Invoice.Status = InvoiceFinalized
	return nil
}

// Unused methods for this specific test, but required by interface
func (m *MockFinalizerStore) CreateInvoice(ctx context.Context, inv *Invoice) error { return nil }
func (m *MockFinalizerStore) GetInvoice(ctx context.Context, t uuid.UUID, y, mo int) (*Invoice, error) {
	return nil, nil
}
func (m *MockFinalizerStore) DeleteInvoice(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockFinalizerStore) UpdateStatus(ctx context.Context, id uuid.UUID, s InvoiceStatus) error {
	return nil
}
func (m *MockFinalizerStore) MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID, transactionID string) error {
	return nil
}

// --- TESTS ---

func TestFinalizeInvoice(t *testing.T) {
	testID := uuid.New()

	tests := []struct {
		name          string
		initialState  *Invoice      // What's in the DB before we start?
		mockFetchErr  error         // Simulate DB error on fetch?
		expectedError error         // What error do we expect?
		verifyStatus  InvoiceStatus // What should the status be after?
	}{
		{
			name: "Happy Path: Draft -> Finalized",
			initialState: &Invoice{
				InvoiceID:  testID,
				Status:     InvoiceDraft,
				TotalCents: 1000,
				Currency:   "USD",
			},
			expectedError: nil,
			verifyStatus:  InvoiceFinalized,
		},
		{
			name: "Idempotency: Already Finalized -> Error",
			initialState: &Invoice{
				InvoiceID:  testID,
				Status:     InvoiceFinalized, // Already done!
				TotalCents: 1000,
				Currency:   "USD",
			},
			expectedError: ErrInvoiceAlreadyFinalized,
			verifyStatus:  InvoiceFinalized, // Should remain unchanged
		},
		{
			name: "Invalid State: Cannot Finalize a VOID Invoice",
			initialState: &Invoice{
				InvoiceID:  testID,
				Status:     InvoiceVoid,
				TotalCents: 1000,
				Currency:   "USD",
			},
			// Expect wrapped error, so we check using errors.Is later usually,
			// but here our mock returns specific error.
			expectedError: ErrInvoiceNotDraft,
			verifyStatus:  InvoiceVoid,
		},
		{
			name: "Integrity Check: Negative Total",
			initialState: &Invoice{
				InvoiceID:  testID,
				Status:     InvoiceDraft,
				TotalCents: -500, // Invalid!
				Currency:   "USD",
			},
			// We expect a text error from fmt.Errorf, checking for nil is basic.
			// Ideally we'd check error string or custom error type.
			expectedError: errors.New("integrity violation: negative total amount -500"),
			verifyStatus:  InvoiceDraft, // Should NOT change
		},
		{
			name: "Integrity Check: Missing Currency",
			initialState: &Invoice{
				InvoiceID:  testID,
				Status:     InvoiceDraft,
				TotalCents: 1000,
				Currency:   "", // Invalid!
			},
			expectedError: errors.New("integrity violation: missing currency"),
			verifyStatus:  InvoiceDraft,
		},
		{
			name:          "Not Found: Invoice Does Not Exist",
			initialState:  nil, // DB is empty
			expectedError: ErrInvoiceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Setup
			store := &MockFinalizerStore{
				Invoice:  tt.initialState,
				FetchErr: tt.mockFetchErr,
			}
			finalizer := NewInvoiceFinalizer(store)

			// 2. Execute
			err := finalizer.FinalizeInvoice(context.Background(), testID)

			// 3. Verify Error
			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectedError)
				}
				// For simple string errors or predefined vars:
				// If expected is a variable (ErrInvoiceAlreadyFinalized), use errors.Is
				if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					// We check if the error *contains* the expected message for Integrity checks
					// Or strictly matches for typed errors
					t.Errorf("Expected error '%v', got '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			// 4. Verify Final State (Immutability Check)
			if tt.initialState != nil {
				// We check the "Database" to see if it changed
				if store.Invoice.Status != tt.verifyStatus {
					t.Errorf("Status mismatch. Expected %s, got %s", tt.verifyStatus, store.Invoice.Status)
				}
			}
		})
	}
}
