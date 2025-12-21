-- services/billing-service/db/migrations/001_create_create_usage_aggregates_table.sql


-- This is the database table where Aggregated usages data will be saved durably in postgres 

CREATE TABLE IF NOT EXISTS usage_aggregates (
    tenant_id UUID NOT NULL,
    usage_type TEXT NOT NULL,
    billing_year INT NOT NULL CHECK (billing_year >= 2000),
    billing_month INT NOT NULL CHECK (billing_month >= 1 AND billing_month <= 12),
    total_quantity BIGINT NOT NULL CHECK (total_quantity >= 0),
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (tenant_id, usage_type, billing_year, billing_month)
);


-- Index for fast billing queries by tenant and period
CREATE INDEX IF NOT EXISTS idx_usage_aggregates_tenant_period
ON usage_aggregates (tenant_id, billing_year, billing_month);

CREATE INDEX IF NOT EXISTS idx_usage_aggregates_last_updated
ON usage_aggregates (last_updated DESC);