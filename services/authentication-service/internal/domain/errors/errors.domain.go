// services/authentication-service/internal/domain/errors/errors.domain.go
package errors

import "errors"

// Standard Sentinel Errors
// These allow the transport layer (gRPC/HTTP) to map internal logic
// to status codes (e.g., ErrInvalidCredentials -> 401 Unauthenticated).

var (
	// Authentication Errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserSuspended      = errors.New("user account is suspended")
	ErrUserDeleted        = errors.New("user account is deleted")
	ErrEmailAlreadyExists = errors.New("email already exists")

	// Authorization/Tenant Errors
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantSuspended     = errors.New("tenant is suspended")
	ErrNotTenantOwner      = errors.New("operation requires tenant ownership")
	ErrMembershipNotFound  = errors.New("membership not found")
	ErrDuplicateMembership = errors.New("user is already a member of this tenant")

	// System/Validation Errors
	ErrInvalidInput        = errors.New("invalid input arguments")
	ErrInternalServerError = errors.New("internal server error")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrNotTenantAdmin      = errors.New("operation requires tenant admin privileges")
	ErrInvalidState        = errors.New("Invalid State")
)
