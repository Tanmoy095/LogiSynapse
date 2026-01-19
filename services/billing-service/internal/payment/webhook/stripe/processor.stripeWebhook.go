// services/billing-service/internal/payment/webhook/stripe/processor.stripeWebhook.go
package stripe

import (
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"

	// Import the core domain
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/payment"
)

type Processor struct {
	secret string
}

func New(secret string) *Processor {
	return &Processor{secret: secret}
}

func (p *Processor) Provider() string {
	return "Stripe"
}

func (p *Processor) VerifyAndParse(payload []byte, headers map[string]string) (*payment.NormalizedEvent, error) {
	// 1. Verify Signature (Security)
	event, err := webhook.ConstructEvent(
		payload,
		headers["Stripe-Signature"],
		p.secret,
	)
	if err != nil {
		return nil, fmt.Errorf("stripe signature invalid: %w", err)
	}

	// 2. Parse JSON
	var pi stripe.PaymentIntent
	// We only care about PaymentIntent objects for now
	if event.Data.Object != nil {
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			// Not a critical error, might be an event type we don't care about structure of
			return nil, nil
		}
	}

	// 3. Map to Domain Event
	switch event.Type {
	case "payment_intent.succeeded":
		return &payment.NormalizedEvent{
			Provider:          "Stripe",
			ProviderPaymentID: pi.ID,
			Status:            payment.PaymentSucceeded,
		}, nil

	case "payment_intent.payment_failed":
		var code, msg *string
		if pi.LastPaymentError != nil {
			c := string(pi.LastPaymentError.Code)
			m := pi.LastPaymentError.Msg
			code, msg = &c, &m
		}
		return &payment.NormalizedEvent{
			Provider:          "Stripe",
			ProviderPaymentID: pi.ID,
			Status:            payment.PaymentFailed,
			ErrorCode:         code,
			ErrorMessage:      msg,
		}, nil
	}

	// Return nil, nil for events we ignore (like "charge.refunded" for now)
	return nil, nil
}
