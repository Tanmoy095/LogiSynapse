// services/billing-service/internal/billing/billing_calculator.go
package billing

import (
	"context"
	"fmt"
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/store"
)

// BillingCalculator orchestrates the transformation
// of Usage -> Money -> Ledger.
type billingCalculator struct {
	usageStore   store.UsageStore
	pricingStore store.PricingStore
	ledgerStore  store.LedgerStore
	// clock allows us to mock time in tests (e.g. setting "now" to specific billing dates)
	clock func() time.Time
}

func NewBillingCalculator(
	usageStore store.UsageStore,
	pricingStore store.PricingStore,
	ledgerStore store.LedgerStore,
) *billingCalculator {
	return &billingCalculator{
		usageStore:   usageStore,
		pricingStore: pricingStore,
		ledgerStore:  ledgerStore,
		clock:        time.Now,
	}
}

// BillPeriod processes all usage for a specific month and generates ledger entries.
// It is idempotent: running it multiple times for the same period is safe.
func (bc *billingCalculator) BillPeriod(ctx context.Context, year int, month int) error {
	// 1. Define the billing timestamp.
	// We usually look up prices effective as of the FIRST day of the billing month.
	billTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC) // means 1st day of month at 00:00 UTC
	fmt.Printf("🔄 Starting billing run for %04d-%02d...\n", year, month)

	// Fetch all aggregated usage for this period
	usageRecords, err := bc.usageStore.GetUsageForPeriod(ctx, year, month)
	if err != nil {
		return fmt.Errorf("failed to get usage for period: %w", err)
	}
	successfulBills := 0

	// for each usage record, calculate cost and create ledger entry
	for _, record := range usageRecords {
		if err := bc.ProcessingSingleRecord(ctx, record, billTime, year, month); err != nil {

			// DECISION: Do we stop the whole run if one tenant fails?
			// usually, we log error and continue, but for strict consistency in this phase,
			// we will return the error to alert the operator.
			return fmt.Errorf("billing failed for tenant %s usage %s: %w", record.TenantID, record.UsageType, err)
		}
		successfulBills++
	}
	fmt.Printf("✅ Billing run complete. Processed %d records.\n", successfulBills)
	return nil

}

func (bc *billingCalculator) ProcessingSingleRecord(
	ctx context.Context,
	record store.UsageRecord,
	billTime time.Time,
	year int,
	month int,
) error {

	// Implementation for processing a single record goes here

	//Fetch the price rule for this usage type and tenant
	priceRule, err := bc.pricingStore.GetPriceRules(ctx, record.UsageType, record.TenantID, billTime)
	if err != nil {
		// If no price exists, we CANNOT bill. This is a critical configuration error.
		return fmt.Errorf("failed to get price rule: %w", err)
	}
	//Calculate Cost (Flat pricing for now)
	// Math: Quantity * UnitPriceCents
	totalCostCents := record.TotalQuantity * priceRule.UnitPriceCents
	if totalCostCents == 0 {
		// Optimization: Don't clutter ledger with $0.00 entries (unless required for audit)
		return nil
	}

	//Create Deterministic Ledger Entry ID
	// Format: "usage_TENANT_YEAR_MONTH_TYP
	ledgerEntryID := fmt.Sprintf("usage_%s_%04d_%02d_%s", record.TenantID.String(), year, month, record.UsageType)
	ledgerEntry := store.LedgerEntry{
		EntryID:         ledgerEntryID,
		TenantID:        record.TenantID,
		TransactionType: string(billingtypes.TransactionTypeDebit), // Customer owes us money..ENFORCED: It's a Debit
		AmountCents:     totalCostCents,
		Currency:        priceRule.Currency,
		// Helpful description for the invoice UI later
		Description: fmt.Sprintf("%s Fee: %d units @ %s %d cents/unit",
			record.UsageType, record.TotalQuantity, priceRule.Currency, priceRule.UnitPriceCents),
	}
	// Step 4: Persist to Ledger
	// The store implementation handles idempotency (ON CONFLICT DO NOTHING

	if err := bc.ledgerStore.CreateLedgerEntry(ctx, ledgerEntry); err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}
	return nil
}
