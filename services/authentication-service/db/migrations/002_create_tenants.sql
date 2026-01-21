--services/authentication-service/db/migrations/002_create_tenants.sql

-- Creates tenants table for isolation boundaries.
-- Invariants: Must have an owner, status management.

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Immutable identifier.
    name TEXT NOT NULL, -- Tenant name.-- Human-readable name (e.g., 'Acme Corp').
    status TEXT NOT NULL DEFAULT 'active',          -- 'active', 'suspended'.
    owner_user_id UUID NOT NULL,                    -- References users.id. Exactly one owner.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Enforce status.
ALTER TABLE tenants ADD CONSTRAINT check_tenant_status CHECK (status IN ('active', 'suspended'));

-- FK to users (invariant: owner must exist; cascading delete optional—app handles).

ALTER TABLE tenants ADD CONSTRAINT fk_tenant_owner FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE RESTRICT;
--ON DELETE RESTRICT: This is a safety lock. If you try to delete a User 
--who still owns a Tenant, the database will stop you and throw an error. 
-- It prevents "orphaned" tenants (tenants with no owner).


-- Indexes:
-- 1. Owner lookup: For queries like "list my tenants" (O(log n)).
CREATE INDEX IF NOT EXISTS idx_tenants_owner ON tenants (owner_user_id);
-- 2. Status filter: Common for active tenants only.
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants (status);

-- Updated_at trigger (same as users).
CREATE TRIGGER trig_tenants_updated_at
BEFORE UPDATE ON tenants
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();
