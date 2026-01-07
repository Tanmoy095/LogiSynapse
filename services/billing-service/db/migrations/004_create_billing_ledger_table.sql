-- services/billing-service/db/migrations/004_create_billing_ledger_table.sql
-- This migration creates the billing_ledger table to store all billing transactions for tenants.
CREATE TABLE billing_ledger (
    id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,


     
     --Audit Trial (Where did this charge came from )
    transaction_type TEXT NOT NULL, -- e.g., "USAGE_FEE", "SUBSCRIPTION_FEE" , "CRREDIT"
    reference_id UUID NOT NULL, --  "usage_2024_06_shipment_created" (IDEMPOTENCY KEY   )

    --THE Financials
    amount_cents BIGINT NOT NULL CHECK , -- Amount Positive = User owes us . negAtive = we owes user.. owe means we have to pay or collect
    quantity BIGINT NOT NULL DEFAULT 1, -- e.g., number of shipments
    unit_price_cents BIGINT NOT NULL DEFAULT 0, -- e.g., price per shipment in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD', -- Currency code (e.g., USD, EUR
    description TEXT, -- eg. 150shipment@0.10 per shipment
    usage_type TEXT NOT NULL, -- e.g., "SHIPMENT_CREATED"
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    
    --Prevent Charges for the same reference_id for the same tenant (Idempotency at DB level
    UNIQUE (tenant_id, reference_id) -- means we cannot have two charges for the same usage event

);