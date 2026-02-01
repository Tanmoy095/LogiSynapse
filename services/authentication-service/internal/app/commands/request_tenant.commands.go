// services/authentication-service/internal/app/commands/request_tenant.commands.go
package commands

import (
	"context"
	"time"

	domainErr "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/tenant"
	"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
	"github.com/google/uuid"
)

//One pending request per user.

type ReqTenantCmd struct {
	userRepo   repository.UserStore
	reqTntRepo repository.TenantRequestStore
}

func NewReqTenantCmd(
	userRepo repository.UserStore,
	reqTntRepo repository.TenantRequestStore,
) *ReqTenantCmd {
	return &ReqTenantCmd{
		userRepo:   userRepo,
		reqTntRepo: reqTntRepo,
	}
}

type reqTntParams struct {
	actorUserID    uuid.UUID
	DesiredTntName string
}

// Implement the command methods here (e.g., Execute)

func (cmd *ReqTenantCmd) Handle(ctx context.Context, params reqTntParams) (uuid.UUID, error) {
	// Check if user exists. Validate User Exists & Active
	user, err := cmd.userRepo.GetUserByID(ctx, params.actorUserID)
	if err != nil || user == nil {
		return uuid.Nil, domainErr.ErrUserNotFound
	}
	if user.Status != "active" {
		return uuid.Nil, domainErr.ErrUserNotActive
	}
	// Anti-Spam Check (Optimization: O(1) Lookup)
	// "One pending request per user"
	existingReq, _ := cmd.reqTntRepo.GetPendingTntRequestByUser(ctx, params.actorUserID)
	if existingReq != nil {
		return uuid.Nil, domainErr.ErrRequestAlreadyPending
	}

	// Create Tenant Request
	newReq := &tenant.TenantCreationRequest{
		ID:                uuid.New(),
		RequesterUserID:   params.actorUserID,
		DesiredTenantName: params.DesiredTntName,
		TenantStatus:      tenant.RequestStatusPending,
		CreatedAt:         time.Now().UTC(),
	}
	if err := cmd.reqTntRepo.CreateTntRequest(ctx, newReq); err != nil {
		return uuid.Nil, err
	}
	return newReq.ID, nil

}
