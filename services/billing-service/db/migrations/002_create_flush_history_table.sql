-- services/billing-service/db/migrations/002_create_flush_history_table.sql

--Table to log batch ids primary key ensures uniqueness

--During flush try INSERT .. if conflict skip upserts  

--Idempotency at db level --safe retries without double adds to quantities  

--Real Example : ""uuid-123" inserted on first flush retry detects conflict- no operation 


CREATE TABLE IF NOT EXISTS flush_history (
    batch_id UUID PRIMARY KEY,
    flushed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

