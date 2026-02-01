// services/authentication-service/internal/domain/tenant/tenant_request.domain.go
package tenant

import (
	"time"

	"github.com/google/uuid"
)

type RequestStatus string

const (
	RequestStatusPending  RequestStatus = "pending"
	RequestStatusApproved RequestStatus = "approved"
	RequestStatusRejected RequestStatus = "rejected"
)

type TenantCreationRequest struct {
	ID                uuid.UUID
	RequesterUserID   uuid.UUID
	DesiredTenantName string
	TenantStatus      RequestStatus
	ReviewerUserID    *uuid.UUID // Nullable
	RejectionReason   *string    // Nullable.. why did we say no?
	CreatedAt         time.Time
	ReviewedAt        *time.Time // Nullable
}

func (request *TenantCreationRequest) Approve(reviewerID uuid.UUID) {
	if request.TenantStatus != RequestStatusPending {
		// In domain logic, we don't return errors for impossible states usually,
		// we assume the command layer checks, but defensive coding is good.
		return
	}
	now := time.Now().UTC()
	request.TenantStatus = RequestStatusApproved
	request.ReviewerUserID = &reviewerID
	request.ReviewedAt = &now
}
func (request *TenantCreationRequest) Reject(reviewerID uuid.UUID, reason string) {
	if request.TenantStatus != RequestStatusPending {
		return
	}
	now := time.Now().UTC()
	request.TenantStatus = RequestStatusRejected
	request.ReviewerUserID = &reviewerID
	request.RejectionReason = &reason
	request.ReviewedAt = &now
}

//*Why approve/reject methods live in DOMAIN? “Why do we have Approve / Reject methods?”

/* Because state transitions must be guarded. Without these methods:

Anyone could set status = approved
Logic would be duplicated
State rules would leak into application layer
This is a finite state machine:
PENDING
   ├── approve → APPROVED
   └── reject  → REJECTED
Once approved/rejected:
❌ Cannot change again, ❌ Cannot be edited. That invariant belongs in domain, not commands.

That invariant belongs in domain, not commands.*/
