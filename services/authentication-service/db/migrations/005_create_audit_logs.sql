--- services/authentication-service/db/migrations/005_create_audit_logs.sql



CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- Who performed the action
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,    -- Context of the action
    action TEXT NOT NULL,                                       -- e.g., 'MEMBER_INVITED', 'TENANT_SUSPENDED'
    target_id UUID,                                             -- The ID of the affected resource
    ip_address INET,                                            -- Securely stores IPv4 or IPv6
    metadata JSONB,                                             -- Flexible context (old/new values)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_at ON audit_logs (tenant_id, created_at DESC);