-- services/billing-service/db/migrations/009_add_ledger_metadata.sql
-- Add metadata columns to billing_ledger in a safe, idempotent way.

ALTER TABLE billing_ledger
  ADD COLUMN IF NOT EXISTS quantity BIGINT NOT NULL DEFAULT 1;

ALTER TABLE billing_ledger
  ADD COLUMN IF NOT EXISTS unit_price_cents BIGINT NOT NULL DEFAULT 0;

-- Optional: ensure amount_cents is present and non-null (do not alter type if DB already has it).
-- If you need to add a CHECK constraint for non-null/valid ranges, do so in a separate, careful migration.
