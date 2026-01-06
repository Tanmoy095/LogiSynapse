//services/billing-service/internal/billing/billing_calculator_test.go

package billing

// --- MOCKS ---
// We use 'testify/mock' pattern (standard in Go pro environments)
// If you don't have testify, run: go get github.com/stretchr/testify/

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/ledger"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/pricing"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/store"
	"github.com/google/uuid"
)

// --- MOCK STORES (Boilerplate for testing) ---

type MockUsageStore struct {
	Records []store.UsageRecord
}

func (m *MockUsageStore) Flush(ctx context.Context, batch store.FlushBatch) error { return nil }
func (m *MockUsageStore) GetUsageForPeriod(ctx context.Context, year, month int) ([]store.UsageRecord, error) {
	return m.Records, nil
}

type MockPricingStore struct {
	Rule pricing.PriceRule
	Err  error
}

func (m *MockPricingStore) GetPriceRules(ctx context.Context, u billingtypes.UsageType, t uuid.UUID, at time.Time) (pricing.PriceRule, error) {
	return m.Rule, m.Err
}

type MockLedgerStore struct {
	Entries []ledger.LedgerEntry
	// We simulate idempotency: if EntryID exists, we don't duplicate
}

func (m *MockLedgerStore) CreateLedgerEntry(ctx context.Context, entry ledger.LedgerEntry) error {
	for _, e := range m.Entries {
		if e.EntryID == entry.EntryID {
			// Idempotency: Simulate "ON CONFLICT DO NOTHING" -> Success (nil error)
			return nil
		}
	}
	m.Entries = append(m.Entries, entry)
	return nil
}

func (m *MockLedgerStore) GetEntriesForPeriod(ctx context.Context, tenantID uuid.UUID, year, month int) ([]ledger.LedgerEntry, error) {
	// For tests we return all stored entries; filter by tenant/year/month if needed in future.
	return m.Entries, nil
}

// --- TESTS ---

func TestBillingCalculator_HappyPath(t *testing.T) {
	// 1. SETUP
	tenantID := uuid.New()
	usageRepo := &MockUsageStore{
		Records: []store.UsageRecord{
			{
				TenantID:      tenantID,
				UsageType:     billingtypes.ShipmentCreated,
				TotalQuantity: 10,
				BillingPeriod: store.BillingPeriod{Year: 2024, Month: 6},
			},
		},
	}
	pricingRepo := &MockPricingStore{
		Rule: pricing.PriceRule{
			UnitPriceCents: 50, // $0.50
			Currency:       "USD",
		},
	}
	ledgerRepo := &MockLedgerStore{}

	calc := NewBillingCalculator(usageRepo, pricingRepo, ledgerRepo)

	// 2. EXECUTE
	err := calc.BillPeriod(context.Background(), 2024, 6)

	// 3. ASSERT
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if len(ledgerRepo.Entries) != 1 {
		t.Fatalf("Expected 1 ledger entry, got %d", len(ledgerRepo.Entries))
	}

	entry := ledgerRepo.Entries[0]
	expectedAmount := int64(10 * 50) // 500 cents
	if entry.AmountCents != expectedAmount {
		t.Errorf("Wrong amount: expected %d, got %d", expectedAmount, entry.AmountCents)
	}
	if entry.TransactionType != billingtypes.TransactionTypeDebit {
		t.Errorf("Wrong type: expected DEBIT, got %s", entry.TransactionType)
	}
	// Check Idempotency Key construction
	expectedID := fmt.Sprintf("usage_%s_2024_06_%s", tenantID.String(), billingtypes.ShipmentCreated)
	if entry.EntryID != expectedID {
		t.Errorf("Wrong EntryID: %s", entry.EntryID)
	}
}

func TestBillingCalculator_Idempotency(t *testing.T) {
	// 1. SETUP
	tenantID := uuid.New()
	usageRepo := &MockUsageStore{
		Records: []store.UsageRecord{
			{TenantID: tenantID, UsageType: billingtypes.APIRequest, TotalQuantity: 100, BillingPeriod: store.BillingPeriod{2024, 6}},
		},
	}
	pricingRepo := &MockPricingStore{Rule: pricing.PriceRule{UnitPriceCents: 1, Currency: "USD"}}
	ledgerRepo := &MockLedgerStore{} // Starts empty

	calc := NewBillingCalculator(usageRepo, pricingRepo, ledgerRepo)

	// 2. EXECUTE TWICE
	_ = calc.BillPeriod(context.Background(), 2024, 6)
	_ = calc.BillPeriod(context.Background(), 2024, 6)

	// 3. ASSERT
	if len(ledgerRepo.Entries) != 1 {
		t.Errorf("Idempotency failed! Expected 1 entry, got %d", len(ledgerRepo.Entries))
	}
}

func TestBillingCalculator_MissingPriceRule(t *testing.T) {
	// 1. SETUP
	usageRepo := &MockUsageStore{
		Records: []store.UsageRecord{
			{TenantID: uuid.New(), UsageType: billingtypes.ShipmentCreated, TotalQuantity: 5},
		},
	}
	// Simulate "Price Not Found"
	pricingRepo := &MockPricingStore{
		Err: errors.New("not found"),
	}
	ledgerRepo := &MockLedgerStore{}

	calc := NewBillingCalculator(usageRepo, pricingRepo, ledgerRepo)

	// 2. EXECUTE
	err := calc.BillPeriod(context.Background(), 2024, 6)

	// 3. ASSERT
	if err == nil {
		t.Fatal("Expected error due to missing price, got nil")
	}
	if len(ledgerRepo.Entries) != 0 {
		t.Error("Should not create ledger entries if pricing fails")
	}
}
