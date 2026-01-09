phase 1 is the foundational step in building oour billing service , focusing exclusively on the core domain service likea database , messege queue or payment gateways ,,

the goal is to established bussiness logic

accounts. folder --->> this handles who pays. example are Subscribers

leadger ---> A leadger is like a financial diary .. every time money (changes, refund and paayment) you write down a note

# phase 2

Customer creates shipments like created shipment (usages type :- "CREATE SHIPMENT )OR make api request (usage type: "API_REQUEST") ..Each action sends an event

**Event** :--> "Tenant ABC created 1 shipment at 25-02-25 10:12 pm"
**Aggregation**--> Sum shipments for ABC this month --> e.g. 500 shipments ->bill $50 if $0.10/shipment

If 10,000 events arrive simultaneously (e.g., Black Friday rush), the code processes them in parallel without messing up counts (no races where two updates overwrite each other).

**Usages.models.go** – Defining What We're Aggregating
This file sets up the basic data types. It's like the "vocabulary" of the system.

UsageType: An enum (string constants) for categories of usage. E.g., ShipmentCreated for shipments, ApiRequest for API calls. Why? Prevents typos and makes code readable—only valid types allowed.
UsageEvent: The input struct—each event is one "usage incident".
Fields:
ID: Unique string (e.g., "event-123") for idempotency (dedup duplicates later).
TenantID: UUID (e.g., "uuid-for-company-ABC")—who owns this usage.
Type: The metric (e.g., ShipmentCreated).
Quantity: How much (e.g., 1 for one shipment, 5 for batch).
Timestamp: When it happened (e.g., time.Now())—for future time-based grouping (e.g., monthly totals).

What's happening: This is the raw data incoming. In the example, an event might be:
**{ID: "event-123", TenantID: "ABC", Type: "SHIPMENT_CREATED", Quantity: 1, Timestamp: 2025-12-19 10:00}. We're aggregating Quantity sums per TenantID + Type.**

**Why aggregate this?** To turn raw events into billable totals. Without it, you'd query millions of events for each invoice—slow and error-prone.

**aggregator.go – The Engine Managing Workers and Events**

This is the "brain"—it handles incoming events, distributes them to workers, and ensures safe aggregation.

phase 2.1 adds durability to pjase 2s aggretes making them usabkr by ohase 1s priceing .
Aggregator/buckets hold temporary sums

WHY STORE PRICING RULES

historical prices for each usages type

🟣 PHASE 3 — CUSTOMER BILLING & PAYMENTS
Phase 3 turns your ledger into real money flow
🔹 Phase 3.1 — Invoice Generation (NO Stripe yet)
Goal
Convert ledger entries → monthly invoices
What you will build
invoice table
Invoice generator
Invoice totals from ledger
Invoice status: DRAFT, FINALIZED, PAID
Why this comes first
Stripe never calculates money.
Stripe only collects money.
🔹 Phase 3.2 — Tiered Pricing Integration
Goal
Activate your pricing_engine
Work
Extend pricing_rules to include tiers (JSON or normalized)
Build engine from rules
Replace flat pricing in BillingCalculator
🔹 Phase 3.3 — Stripe Integration (LAST)
Goal
Collect payment for finalized invoices
Work
Stripe Customer mapping
PaymentIntent creation
Webhooks → update invoice & ledger
Handle retries safely
🔹 Phase 3.4 — Billing API (Read-Only First)
Expose:
Usage summary
Invoice history
Ledger view
5️⃣ INTERVIEW TRUTH (IMPORTANT)
If you explain just Phase 2.2 like this in an interview:
Deterministic billing
Idempotent ledger
Retry-safe aggregation
Pricing vs billing separation
👉 You clear senior backend interviews.
Very few candidates can reason at this level

This is a High-Impact Architectural Fix.

You identified the flaw correctly: The Ledger was storing the "Result" (Total $) but losing the "Context" (Quantity & Price). An invoice line that says "Total: $50.00" without saying "100 Units @ $0.50" is legally insufficient in many countries and confusing for users.

To fix this professionally, we must strictly follow the data flow: Usage (Source) → Billing Calculator → Ledger (Storage) → Invoice (Display)

The Ledger must store this metadata.

🛠️ Step 1: Database Migration (Schema Fix)

We need to add columns to the billing_ledger. If you are still in early development and haven't deployed to production, you can edit migration 004 directly. If you want to be safe, create migration 006.

Option A: Edit 004_create_billing_ledger_table.sql (Cleaner for new projects)

SQL
CREATE TABLE billing_ledger (
-- ... existing fields ...
amount_cents BIGINT NOT NULL,

    -- ✅ NEW METADATA COLUMNS
    quantity BIGINT NOT NULL DEFAULT 1,       -- How many units?
    unit_price_cents BIGINT NOT NULL DEFAULT 0, -- Price per unit?

    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    -- ... existing fields ...

);
Option B: New Migration 006_add_ledger_metadata.sql (Safe)

SQL
ALTER TABLE billing_ledger
ADD COLUMN quantity BIGINT NOT NULL DEFAULT 1,
ADD COLUMN unit_price_cents BIGINT NOT NULL DEFAULT 0;
📦 Step 2: Update Ledger Model & Store

We need to carry this data through the Go structs.

File: internal/ledger/models.ledger.go

Go
type LedgerEntry struct {
EntryID string
TenantID uuid.UUID
TransactionType TransactionType
AmountCents int64
Currency string
Description string
CreatedAt time.Time
UsageType billingtypes.UsageType

    // ✅ NEW FIELDS
    Quantity       int64
    UnitPriceCents int64

}
File: internal/store/postgres/ledger_store.postgres.go Update the INSERT and SELECT queries to include these new fields.

Go
// In CreateLedgerEntry
query := `  INSERT INTO billing_ledger 
  (tenant_id, transaction_type, reference_id, amount_cents, usage_type, currency, description, quantity, unit_price_cents, created_at)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
  ON CONFLICT (tenant_id, reference_id) DO NOTHING;`
// Pass entry.Quantity and entry.UnitPriceCents as $8 and $9

// In GetEntriesForPeriod
// Add quantity, unit_price_cents to the SELECT list and Scan()
⚙️ Step 3: Fix Billing Calculator (Phase 2.2)

Now we must populate these fields when we calculate the bill.

File: internal/billing/billing_calculator.go

Go
func (bc \*billingCalculator) ProcessingSingleRecord(...) error {
// ... fetch price rule ...

    totalCostCents := record.TotalQuantity * priceRule.UnitPriceCents

    ledgerEntry := ledger.LedgerEntry{
        // ... existing fields ...
        AmountCents:     totalCostCents,

        // ✅ POPULATE METADATA
        Quantity:       record.TotalQuantity,
        UnitPriceCents: priceRule.UnitPriceCents,

        Description: fmt.Sprintf("%s Fee", record.UsageType), // Keep description simple now, metadata has the details
    }

    return bc.ledgerStore.CreateLedgerEntry(ctx, ledgerEntry)

}
📑 Step 4: Fix Invoice Generator (Phase 3.1)

This is where we solve the "Unit Price is Zero" and "Naive Currency" issues.

File: internal/invoice/invoice_generator.go

Go
func (ig *InvoiceGenerator) GenerateInvoiceForTenant(...) (*Invoice, error) {
// ... fetch existing invoice ...
// ... fetch ledger entries ...

    lineMap := make(map[billingtypes.UsageType]*InvoiceLine)
    var invoiceTotal int64

    // ✅ CURRENCY GUARD
    // We expect the first entry to dictate the currency.
    // In a multi-currency system, you'd group by currency, but here we enforce consistency.
    expectedCurrency := "USD"
    if len(entries) > 0 {
        expectedCurrency = entries[0].Currency
    }

    for _, entry := range entries {
        // 1. Currency Check
        if entry.Currency != expectedCurrency {
            // TODO: Handle multi-currency gracefully. For now, fail safe.
            return nil, fmt.Errorf("currency mismatch in ledger: found %s expected %s", entry.Currency, expectedCurrency)
        }

        var amount int64
        if entry.TransactionType == ledger.Debit {
            amount = entry.AmountCents
        } else if entry.TransactionType == ledger.Credit {
            amount = -entry.AmountCents
        } else {
            continue
        }

        invoiceTotal += amount

        // 2. Grouping Logic with Metadata
        if line, exists := lineMap[entry.UsageType]; exists {
            line.LineTotalCents += amount

            // ✅ Sum Quantity for the line
            line.Quantity += entry.Quantity

            // Note: UnitPrice should ideally be consistent per usage type.
            // If prices changed mid-month, this simple grouping might show an "Average" or last price.
            // For Phase 3.1, keeping the first UnitPrice is acceptable, or we set to 0 if they differ.
             if line.UnitPriceCents != entry.UnitPriceCents {
                // Complex case: Mixed prices for same usage type (e.g. tiered pricing).
                // Solution: Set UnitPrice to 0 to indicate "Variable Rate" on invoice
                line.UnitPriceCents = 0
            }
        } else {
            lineMap[entry.UsageType] = &InvoiceLine{
                ID:             uuid.New(),
                UsageType:      entry.UsageType,
                Description:    entry.Description, // e.g. "Shipment Created Fee"
                LineTotalCents: amount,

                // ✅ MAP METADATA
                Quantity:       entry.Quantity,
                UnitPriceCents: entry.UnitPriceCents,
            }
        }
    }

    // ... flatten and save ...

}
🧠 Summary of Fixes

DB: Added quantity and unit_price columns to Ledger. The Ledger is now a complete audit trail.

Calculator: Now records how the cost was derived (Qty \* Price) into the ledger.

Generator: Now reads that data to build rich Invoice Lines.

Currency: Added a check to ensure we don't accidentally mix currencies in one invoice.

This makes your system professional. The user will see: "150 Shipments @ $0.50 = $75.00" Instead of: "Shipment Fee: $75.00"

Phase 3.1: The Construction Phase (Aggregation)

"From Chaos to Structure"

In Phase 3.1, we focused on Data Aggregation. The problem was that raw financial data (Ledger Entries) is too granular for humans to understand. A customer doesn't want to see 10,000 individual API calls; they want to see "API Usage: $10.00".

Source of Truth: We strictly enforced that the Ledger is the source of truth. We did not recalculate prices from raw usage (which would be risky); we simply summed up what was already recorded in the Ledger.

The Artifact: We created the Invoice struct (Header) and InvoiceLine structs (Details).

Idempotency: We handled the case where a draft is generated multiple times. We decided that since a Draft is mutable, we can delete the old draft and regenerate it fresh if new data comes in.

Rich Metadata: We fixed a critical flaw where line items were missing context. We ensured that every line item carries Quantity and Unit Price so the invoice is legally compliant and transparent to the user.

Key Technical Achievement: Transforming high-volume transactional rows into a single, human-readable document without losing penny-perfect accuracy.

Phase 3.2: The Legal Phase (Finalization)

"From Draft to Contract"

In Phase 3.2, we shifted focus to State Management and Immutability. A "Draft" is a scratchpad; a "Finalized Invoice" is a legal debt obligation. Once an invoice is finalized, it represents real money owed.

The Gatekeeper: We implemented InvoiceFinalizer. Its job is to say "Stop" if the invoice isn't ready. It checks for data integrity (e.g., "Is the total negative?", "Is the currency missing?").

The State Machine: We enforced a strict one-way door: DRAFT → FINALIZED.

You cannot finalize a VOID invoice.

You cannot finalize an already FINALIZED invoice.

Optimistic Locking: In the database layer, we used the SQL clause WHERE id = $1 AND status = 'DRAFT'. This is a professional pattern that prevents race conditions. If two requests try to finalize the same invoice simultaneously, the database ensures only one succeeds.

Immutability: By locking the status, we effectively froze the invoice. Even if the Ledger changes afterwards (e.g., a late refund), this specific invoice document will never change. This is crucial for audit trails.

Key Technical Achievement: Creating a secure "Sealing" process that guarantees the invoice cannot be tampered with once it becomes a legal request for payment.
