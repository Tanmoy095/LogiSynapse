//services/billing-service/internal/payment/retry_policy.go

package payment

import (
	"errors"
	"net"
	"syscall"

	"github.com/stripe/stripe-go/v79"
)

func IsRetryAbleError(err error) bool {
	if err == nil { // No error, no retry needed
		return false
	}
	// Here we can check for specific error types or codes
	return isRetryAbleStripeError(err) || isRetryAbleNetworkError(err) || isRetryAbleSystemError(err)

}

func isRetryAbleStripeError(err error) bool {

	var stripeError *stripe.Error
	// If the error is not a Stripe error, it is not retryAble by this policy.
	if !errors.As(err, &stripeError) {
		return false
	}
	// HTTP 400-499: Client Error (Card Declined, Invalid Request) -> STOP
	// HTTP 500-599: Server Error (Stripe Down) -> RETRY
	if stripeError.HTTPStatusCode >= 500 && stripeError.HTTPStatusCode < 600 {
		return true
	}
	// Stripe throttling / locking → retry
	switch stripeError.Code {
	case stripe.ErrorCodeRateLimit, //too many requests
		stripe.ErrorCodeLockTimeout:
		return true

		// Card / user errors → NEVER retry

	case stripe.ErrorCodeCardDeclined,
		stripe.ErrorCodeExpiredCard,
		stripe.ErrorCodeIncorrectCVC:
		return false
	}
	// Default to false for other API errors (like "Invalid API Key")
	return false

}
func isRetryAbleNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) { //that means it's a network error
		// Timeout or Temporary Network Blip
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}
	return false
}
func isRetryAbleSystemError(err error) bool {
	//Connection Refused / Reset
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	return false
}
