// services/billing-service/internal/payment/webhook.types.go
package payment

// NormalizedEvent is the "Universal Language" of our payment system.
// It doesn't matter if it came from Stripe, PayPal, or a Bank,
// it always looks like this to our Service.
type NormalizedEvent struct {
	Provider          string        // e.g., "Stripe"
	ProviderPaymentID string        // e.g., "pi_3M..."
	Status            PaymentStatus // e.g., SUCCEEDED, FAILED
	ErrorCode         *string       // e.g., "card_declined"
	ErrorMessage      *string       // e.g., "Insufficient funds"
}

// WebhookProcessor describes a component that can parse raw HTTP bytes
// into our NormalizedEvent.
type WebhookProcessor interface {
	Provider() string
	VerifyAndParse(payload []byte, headers map[string]string) (*NormalizedEvent, error)
}
