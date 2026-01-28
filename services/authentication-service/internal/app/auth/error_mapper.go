package auth

import (
	stdErrors "errors"

	domainErr "github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/domain/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//We must prevent User Enumeration. If an attacker tries to log in
// with admin@example.com and gets "Invalid Password", but tries random@example.com
// and gets "User Not Found", they now know admin@example.com exists.
// This mapper flattens those errors.

//Why app/auth/?
//❌ Not domain → domain should not know about gRPC, HTTP, status codes
//❌ Not transport → transport should not contain business rules
//✅ Application layer → translates domain errors → transport-safe errors

func MapLoginError(err error) error {
	if err == nil {
		return nil
	}
	//  Authentication failures (flattened to prevent enumeration)
	if stdErrors.Is(err, domainErr.ErrInvalidCredentials) || stdErrors.Is(err, domainErr.ErrUserNotFound) {
		return status.Error(codes.Unauthenticated, "invalid email or password")
	}
	// Account state errors (intentional leakage for UX/support)
	if stdErrors.Is(err, domainErr.ErrUserSuspended) {
		return status.Error(codes.PermissionDenied, "account suspended")
	}
	if stdErrors.Is(err, domainErr.ErrUserDeleted) {
		return status.Error(codes.PermissionDenied, "account deleted")
	}

	//  Fallback (never leak internals)
	return status.Error(codes.Internal, "internal error")

}
