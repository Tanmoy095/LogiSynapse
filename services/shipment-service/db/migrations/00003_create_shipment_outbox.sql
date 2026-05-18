-- +goose Up
CREATE TABLE IF NOT EXISTS shipment_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    event_key TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_shipment_outbox_pending
ON shipment_outbox (aggregate_id, created_at)
WHERE published_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS shipment_outbox;
