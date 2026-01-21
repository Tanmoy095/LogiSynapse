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


--uses email TEXT NOT NULL for linking, but this is brittle (emails can change; users might update them in auth). Recommendation (2026-proof): Migrate billing to use auth_user_id UUID NOT NULL (references auth users.id). Add:SQLALTER TABLE accounts ADD COLUMN auth_user_id UUID NOT NULL;
--ALTER TABLE accounts ADD CONSTRAINT fk_accounts_user FOREIGN KEY (auth_user_id) REFERENCES authentication.users(id) ON DELETE CASCADE;
--CREATE INDEX IF NOT EXISTS idx_accounts_auth_user_id ON accounts (auth_user_id);
--Why? Immutable UUIDs > mutable emails. Populate via events (e.g., auth publishes "user.created" to Kafka; billing consumes and creates account with auth_user_id).