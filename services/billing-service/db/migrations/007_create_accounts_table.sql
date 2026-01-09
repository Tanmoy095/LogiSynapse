--- services/billing-service/db/migrations/006_create_accounts_table.sql

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Core Identity (Linked to Auth Service via UUID)
    email TEXT NOT NULL,
    
    -- Subscription State
    current_plan TEXT NOT NULL DEFAULT 'FREE',
    
    -- Stripe Integration
    stripe_customer_id TEXT,    -- "cus_..."
    payment_method_id TEXT,     -- "pm_..."
    
    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_stripe_id ON accounts(stripe_customer_id);