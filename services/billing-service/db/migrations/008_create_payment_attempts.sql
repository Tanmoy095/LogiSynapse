--services/billing-service/db/migrations/008_create_payment_attempts.sql

-- This table implements the "Payment State Machine".
-- It tracks the lifecycle of a payment attempt distinct from the invoice itself.
-- One Invoice can have many Payment Attempts (e.g. failed retries).

CREATE TABLE IF NOT EXISTS payment_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    
   -- Who are we talking to?
   provider TEXT NOT NULL DEFAULT 'stripe', -- e.g., "STRIPE", "PAYPAL"
   -- The External ID (e.g., Stripe PaymentIntent ID). 
    -- Initially NULL, populated once Stripe responds or we generate a key.
    provider_payment_id TEXT,
    -- The State Machine
    -- PENDING: Created locally, waiting for Stripe.
    -- SUCCEEDED: Money captured.
    -- FAILED: Card declined or network error.
    -- REQUIRES_ACTION: 3DSecure/OTP needed (Advanced flow).
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'SUCCEEDED', 'FAILED', 'REQUIRES_ACTION')),
    -- Snapshot of financial data at the time of attempt
    -- (In case invoice changes later, this record remains historical truth)
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,

    -- Error tracking for debugging/support
    error_code TEXT,
    error_message TEXT,
    
    -- Retry Logic (Phase 3.4.5)
    retry_count INT DEFAULT 0,
    next_retry_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Indexes are CRITICAL for the Reconciliation Worker
-- "Find me all pending payments older than 5 minutes"
CREATE INDEX IF NOT EXISTS idx_payment_attempts_status_created 
ON payment_attempts(status, created_at);

-- "Find all attempts for this invoice"
CREATE INDEX IF NOT EXISTS idx_payment_attempts_invoice 
ON payment_attempts(invoice_id);

-- "Find attempt by Stripe ID" (for Webhooks)
CREATE INDEX IF NOT EXISTS idx_payment_attempts_provider_id 
ON payment_attempts(provider_payment_id);