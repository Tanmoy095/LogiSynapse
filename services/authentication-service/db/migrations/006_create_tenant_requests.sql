--services/authentication-service/db/migrations/006_create_tenant_requests.sql

-- Track who ask for a tenant ..

CREATE TABLE IF NOT EXISTS tenant_creation_requests(

id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Immutable identifier.
requester_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- References users.id. if user is deleted ,their requests will be deleted too.
desired_tenant_name TEXT NOT NULL, -- Desired tenant name.
status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'rejected'.
reviewed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- Who reviewed the request. means if user is deleted, this field will be set to NULL.
rejection_reason TEXT, -- Reason for rejection, if applicable.
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
reviewed_at TIMESTAMPTZ,
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--Index for "GetPendingByUserID" (Spam check)
CREATE UNIQUE INDEX IF NOT EXISTS idx_request_pending_user
ON tenant_creation_requests (requester_user_id)
WHERE status = 'pending';
--Enforce Enum values for status
ALTER TABLE tenant_creation_requests
ADD CONSTRAINT check_request_status
CHECK (status IN ('pending', 'approved', 'rejected'));