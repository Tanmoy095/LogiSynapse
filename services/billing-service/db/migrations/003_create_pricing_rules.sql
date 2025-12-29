--services/billing-service/db/migrations/003_create_pricing_rules.sql
-- This migration creates the pricing_rules table to store pricing rules for different usage types per tenant.

CREATE TABLE pricing_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Unique identifier for each pricing rule. ID is critical for referencing specific rules.
    usage_type TEXT NOT NULL,
    tenant_id UUID , -- Tenant to which this pricing rule applies,-- NULL allows for "Default Global Price"
    --The Money
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0), -- Price per unit in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD', -- Currency code (e.g., USD, EUR


    --validation constraints (Critical for historical billing accuracy)
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ, -- NULL means "until forever"

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),


);
--Index to find the active price quickly
CREATE INDEX IF NOT EXISTS idx_pricing_lookup
ON pricing_rules (usage_type, tenant_id, effective_from DESC);