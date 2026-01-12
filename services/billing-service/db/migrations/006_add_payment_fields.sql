--services/billing-service/db/migrations/006_add_payment_fields.sql

-- Safe migration to add payment fields to existing invoices table
ALTER TABLE invoices 
ADD COLUMN IF NOT EXISTS payment_intent_id TEXT, -- Stripe Payment Intent ID
ADD COLUMN IF NOT EXISTS error_message TEXT; -- Good for debugging failed payments

-- Create index for faster lookups (e.g. "Find invoice for this Stripe payment")
CREATE INDEX IF NOT EXISTS idx_invoices_payment_intent ON invoices(payment_intent_id);

