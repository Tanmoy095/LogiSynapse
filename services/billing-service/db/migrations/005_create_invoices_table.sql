-- 1. Invoices Header
CREATE TABLE IF NOT EXISTS invoices(
    invoice_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    billing_year INT NOT NULL,
    billing_month INT NOT NULL,

    --financial details (Derived from lines)
    total_amount_cents BIGINT NOT NULL CHECK (total_amount_cents >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',

    -- Lifecycle Status
    status TEXT NOT NULL CHECK (status IN ('DRAFT', 'FINALIZED', 'PAID', 'VOID')) DEFAULT 'DRAFT',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finalized_at TIMESTAMPTZ, -- Set when status moves to FINALIZED
    paid_at TIMESTAMPTZ,      -- Set when status moves to PAID

    -- Business Rule: Only one invoice per tenant per month
    UNIQUE (tenant_id, billing_year, billing_month)
)

-- 2. Invoices Lines (The Detail)

CREATE TABLE IF NOT EXISTS invoice_lines (
    line_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(invoice_id) ON DELETE CASCADE, -- Foreign Key to Invoices Table.. on delete cascade means if invoice is deleted, its lines are too
    usage_type TEXT NOT NULL,
    quantity BIGINT NOT NULL ,
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0),
    line_total_cents BIGINT NOT NULL CHECK (line_total_cents >= 0)
    description TEXT NOT NULL,
)
-- Index for fast retrieval of lines by invoice
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_period  --index name explains its purpose
ON invoices (tenant_id, billing_year, billing_month); --means we can quickly find invoices for a tenant in a specific month/year