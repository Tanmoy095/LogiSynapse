--Services/authentication-service/db/migrations/001_create_users.sql

-- Invariants enforced: global email uniqueness (case-normalized), status management.
--Users are foundational—tenants/memberships reference them
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),--Immutable identifier.
    email TEXT NOT NULL, --Case-normalized (lowercase) email. App logic normalizes before insert.
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active', --Enum: 'active', 'DELETED', 'suspended', USE CHECK FOR SAFETY.
    is_super_admin BOOLEAN NOT NULL DEFAULT FALSE, --Global super-admin flag. Global privilege flag, orthogonal to tenants.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- Immutable creation timestamp.
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()   -- Last update timestamp (app updates on changes).
);
-- Enforce status values (using real ENUM if Postgres 15+)
ALTER TABLE users ADD CONSTRAINT check_user_status CHECK (status IN ('active', 'deleted', 'suspended'));
--This is a CHECK Constraint. It acts like a gatekeeper for the status column.


-- Global uniqueness on normalized email (invariant: no duplicates).
ALTER TABLE users ADD CONSTRAINT unique_user_email UNIQUE (email);
-- if  email  is not unique , this constraint will prevent insertion/updates that violate uniqueness.


-- Indexes:
-- 1. Email lookup: Hot path for login/register (O(log n) B-tree).
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
-- 2. Partial index for active users (optimization if queries filter status; O(log n) for common cases).
CREATE INDEX IF NOT EXISTS idx_users_active ON users (id) WHERE status = 'active';

-- Index for searching by name (searching "John Doe")
CREATE INDEX IF NOT EXISTS idx_users_full_name ON users (first_name, last_name);

-- Trigger for updated_at ( auto-update on changes).
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trig_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();