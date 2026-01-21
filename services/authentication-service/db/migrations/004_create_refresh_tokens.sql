--services/authentication-service/db/migrations/004_create_refresh_tokens.sql

-- Stores refresh token families for session persistence.
-- Invariants: Hashed storage, rotation chain, revocation.

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Immutable identifier.-- Token ID.

    user_id UUID NOT NULL,                         -- References users.id.
    tenant_id UUID,                                 -- Optional tenant (for tenant-scoped sessions).
    token_hash TEXT NOT NULL,                      -- Hashed token value.SHA-256 hash of opaque token (NEVER raw).
    family_id UUID NOT NULL,                        -- Rotation chain ID (detect replays).
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,                -- Absolute expiry.
    revoked_at TIMESTAMPTZ,                         -- NULL if active.
    replaced_by_token_id UUID,                      -- Links to new token in rotation.
    device_fingerprint TEXT,                        -- Optional hash for risk (IP/UA).
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);
-- Unique on hash (fast lookup, prevent duplicates).
ALTER TABLE refresh_tokens ADD CONSTRAINT unique_refresh_token_hash UNIQUE (token_hash);

-- FKs.
ALTER TABLE refresh_tokens ADD CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE refresh_tokens ADD CONSTRAINT fk_refresh_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

-- Indexes:
-- 1. Hash lookup: Hot refresh path (O(log n)).
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens (token_hash);
-- 2. By user+tenant+revoked: For list active sessions (O(log n)).
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_tenant_revoked ON refresh_tokens (user_id, tenant_id, revoked_at);
-- 3. By family+revoked: Replay detection (O(log n)).
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family_revoked ON refresh_tokens (family_id, revoked_at);

-- Updated_at trigger.
CREATE TRIGGER trig_refresh_tokens_updated_at
BEFORE UPDATE ON refresh_tokens
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();