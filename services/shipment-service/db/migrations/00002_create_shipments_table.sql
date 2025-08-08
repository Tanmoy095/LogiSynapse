-- +goose Up
-- Enable UUID extension for generating unique IDs
-- Why: Allows PostgreSQL to create unique shipment IDs automatically
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    status TEXT NOT NULL,
    eta TEXT,
    carrier_name TEXT,
    carrier_tracking_url TEXT,                    -- Shippo’s tracking URL (nullable)
    tracking_number TEXT,                         -- Shippo’s tracking number (nullable)
    length DOUBLE PRECISION,                      -- Package length (e.g., 12.0)
    width DOUBLE PRECISION,                       -- Package width (e.g., 8.0)
    height DOUBLE PRECISION,                      -- Package height (e.g., 1.0)
    weight DOUBLE PRECISION,                      -- Package weight (e.g., 0.5)
    unit TEXT                                     -- Unit for dimensions/weight (e.g., "in")
);

-- +goose Down
-- Drop the shipments table and UUID extension
-- Why: Reverses the migration for cleanup or rollback
DROP TABLE IF EXISTS shipments;
DROP EXTENSION IF EXISTS pgcrypto;
