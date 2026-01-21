-- services/billing-service/db/migrations/004_create_billing_ledger_table.sql
-- This migration creates the billing_ledger table to store all billing transactions for tenants.

CREATE TABLE IF NOT EXISTS billing_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- Audit trail
    transaction_type TEXT NOT NULL, -- e.g., 'USAGE_FEE', 'SUBSCRIPTION_FEE', 'PAYMENT'
    reference_id UUID NOT NULL, -- idempotency/reference key (usage or payment id)

    -- Financials
    amount_cents BIGINT NOT NULL, -- Positive = customer owes us, Negative = we owe customer
    quantity BIGINT NOT NULL DEFAULT 1,
    unit_price_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD', -- ISO 4217
    description TEXT,
    usage_type TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Idempotency constraint
    UNIQUE (tenant_id, reference_id)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_billing_ledger_tenant_created_at ON billing_ledger (tenant_id, created_at);
CREATE INDEX IF NOT EXISTS idx_billing_ledger_usage_type ON billing_ledger (usage_type);