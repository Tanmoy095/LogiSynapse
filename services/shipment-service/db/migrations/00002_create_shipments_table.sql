-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    status TEXT NOT NULL,
    eta TEXT,
    carrier_name TEXT,
    carrier_tracking_url TEXT
);

-- +goose Down
DROP TABLE IF EXISTS shipments;
DROP EXTENSION IF EXISTS pgcrypto;
