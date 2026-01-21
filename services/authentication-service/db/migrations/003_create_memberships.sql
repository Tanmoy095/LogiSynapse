--services/authentication-service/db/migrations/003_create_memberships.sql

-- Joins users to tenants with roles.

-- Invariants: Unique per (user, tenant), valid refs.


CREATE TABLE IF NOT EXISTS memberships (
    user_id UUID NOT NULL,    -- References users.id.
    tenant_id UUID NOT NULL,   -- References tenants.id.
    role TEXT NOT NULL DEFAULT 'member',        -- e.g., 'owner', 'admin', 'member' . Enum-like..
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique per pair (invariant: no duplicates).
ALTER TABLE memberships ADD CONSTRAINT unique_membership PRIMARY KEY(user_id,tenant_id);
-- FKs (enforce existence).
ALTER TABLE memberships ADD CONSTRAINT fk_membership_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE; --Delete if user deleted.
ALTER TABLE memberships ADD CONSTRAINT fk_membership_tenant FOREIGN KEY(tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;--Delete if tenant deleted.

-- Enforce role values.

ALTER TABLE memberships ADD CONSTRAINT check_membership_role CHECK (role IN ('owner', 'admin', 'member'));

-- Indexes:
-- 1. By tenant+user: Hot for authz checks (O(log n)).

CREATE INDEX IF NOT EXISTS idx_memberships_tenant_user ON memberships(tenant_id, user_id);

-- 2. By user+tenant: For "my memberships" (O(log n)).
CREATE INDEX IF NOT EXISTS idx_memberships_user_tenant ON memberships (user_id, tenant_id);

-- Enable Row Level Security on sensitive tables
ALTER TABLE memberships ENABLE ROW LEVEL SECURITY;

-- Create a policy: A user can only see memberships if their user_id matches
CREATE POLICY user_membership_policy ON memberships
    FOR SELECT
    USING (user_id = auth.uid()); -- auth.uid() is a helper from your app context

-- Updated_at trigger.
CREATE TRIGGER trig_memberships_updated_at
BEFORE UPDATE ON memberships
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();