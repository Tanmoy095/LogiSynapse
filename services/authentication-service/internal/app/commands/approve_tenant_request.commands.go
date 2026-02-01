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

//services/authentication-service/internal/app/commands/approve_tenant_request.commands.go

type AppTntReqCmd struct {
	reqRepo    repository.TenantRequestStore
	tenantRepo repository.TenantStore
	auditRepo  repository.AuditStore
	tx         repository.TransactionManager
}

func NewAppTntReqCmd(
	reqRepo repository.TenantRequestStore,
	tenantRepo repository.TenantStore,
	auditRepo repository.AuditStore,
	tx repository.TransactionManager,
) *AppTntReqCmd {
	return &AppTntReqCmd{
		reqRepo:    reqRepo,
		tenantRepo: tenantRepo,
		auditRepo:  auditRepo,
		tx:         tx,
	}
}

type ApproveTenantRequestParams struct {
	requestID      uuid.UUID
	approverUserID uuid.UUID
	IsSuperAdmin   bool
}

// Handle approves a pending tenant creation request.
func (cmd *AppTntReqCmd) Handle(ctx context.Context, params ApproveTenantRequestParams) error {
	if !params.IsSuperAdmin {
		return domainErr.ErrUnauthorizedAction
	}

	//start transaction
	return cmd.tx.RunInTx(ctx, func(ctx context.Context) error {
		req, err := cmd.reqRepo.GetTntRequestByID(ctx, params.requestID)
		if err != nil {
			return domainErr.ErrRequestNotFound

		}
		if req.TenantStatus != tenant.RequestStatusPending {
			return domainErr.ErrRequestNotPending // Cannot approve processed requests
		}
		//Domain State Transition. it means changing status from pending to approved
		req.Approve(params.approverUserID)
		//Create Tenant according to request
		newTenant := &tenant.Tenant{
			TenantID:     uuid.New(),
			TenantName:   req.DesiredTenantName,
			TenantStatus: tenant.TenantStatusActive,
			OwnerUserID:  req.RequesterUserID, // The requester becomes the owner
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		//  Persist Changes (Both must succeed)
		if err := cmd.reqRepo.UpdateTntRequest(ctx, req); err != nil {
			return fmt.Errorf("failed to update request: %w", err)
		}
		if err := cmd.tenantRepo.CreateTenantWithOwnership(ctx, newTenant); err != nil {

			return fmt.Errorf("failed to create tenant: %w", err)
		}
		// FUTURE:
		// - Emit TenantCreated event (Kafka)
		// - Send webhook notification
		// - Write audit log

		// 🔔 FUTURE EVENT HOOK
		// cmd.auditRepo.Record(...)
		// cmd.eventBus.Publish(TenantCreated

		// 🔐 AUDIT EVENT
		event := &audit.AuditEvent{
			ID:          uuid.New(),
			ActorUserID: &params.approverUserID,
			TenantID:    &newTenant.TenantID,
			Action:      "TENANT_REQUEST_APPROVED",
			TargetID:    &req.ID,
			Metadata: map[string]any{
				"tenant_name": newTenant.TenantName,
				"owner_id":    newTenant.OwnerUserID,
			},
			CreatedAt: time.Now().UTC(),
		}

		return cmd.auditRepo.Append(ctx, event)

	})

}
