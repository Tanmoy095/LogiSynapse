// services/authentication-service/internal/app/commands/reject_tenant_request.commands.go
package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/audit"
	domainErr "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/tenant"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
	"github.com/google/uuid"
)

type RejectTenantCmd struct {
	reqRepo   repository.TenantRequestStore
	auditRepo repository.AuditStore
}

func NewRejectTenantCmd(reqRepo repository.TenantRequestStore, auditRepo repository.AuditStore) *RejectTenantCmd {
	return &RejectTenantCmd{
		reqRepo:   reqRepo,
		auditRepo: auditRepo,
	}
}

type RejectTenantParams struct {
	requestID      uuid.UUID
	approverUserID uuid.UUID
	IsSuperAdmin   bool
	Reason         string
}

// Handle rejects a pending tenant creation request.
func (cmd *RejectTenantCmd) Handle(ctx context.Context, params RejectTenantParams) error {
	if !params.IsSuperAdmin {
		return domainErr.ErrUnauthorizedAction
	}
	req, err := cmd.reqRepo.GetTntRequestByID(ctx, params.requestID)
	if err != nil {
		return domainErr.ErrRequestNotFound
	}
	if req.TenantStatus != tenant.RequestStatusPending {
		return domainErr.ErrRequestNotPending // Cannot approve processed requests
	}
	req.Reject(params.approverUserID, params.Reason)
	// Persist Changes
	if err := cmd.reqRepo.UpdateTntRequest(ctx, req); err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}
	// FUTURE:
	// - Emit TenantRejected event
	// - Notify requester via email/Slack

	// 🔐 AUDIT EVENT
	event := &audit.AuditEvent{
		ID:          uuid.New(),
		ActorUserID: &params.approverUserID,
		Action:      "TENANT_REQUEST_REJECTED",
		TargetID:    &req.ID,
		Metadata: map[string]any{
			"reason": params.Reason,
		},
		CreatedAt: time.Now().UTC(),
	}

	return cmd.auditRepo.Append(ctx, event)

}
