//services/billing-service/internal/payment/stripe_gateway.go

package payment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/client"
)

//WE will implement thew  ChargeAttempt method for Stripe here
//with strict  "OFF-SESSION" Logic as per Stripe's best practices for handling payments without user interaction.
//e.g. user is asleep  dont ask for 2fa unless absolutely necessary

//StripeGAteway Implement payment gateway interface for Stripe

type StripeGateway struct {
	client *client.API //this is the stripe client . it will be initialized with the secret key
}

// NewSt`ripeGateway creates a new StripeGateway with the provided secret key
func NewStripeGateway(apiKey string) *StripeGateway {
	sc := &client.API{}
	sc.Init(apiKey, nil) //initialize stripe client with api key

	return &StripeGateway{client: sc}

}

// ChargeAttempt executes a synchronous charge .off-session charge attempt. this method will handle 3ds and other authentication automatically as per stripe best practices
func (sg *StripeGateway) ChargeAttempt(ctx context.Context, chargeReq PaymentRequest) (*PaymentResult, error) {

	//Input Validation

	if chargeReq.AmountCents <= 0 {
		return nil, ErrInvalidAmount
	}
	if chargeReq.PaymentMethodID == "" || chargeReq.CustomerID == "" {
		return nil, ErrNoPaymentMethod
	}
	// Payment Preparation
	//we map domain  request  to -->> stripe specific parameters
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(chargeReq.AmountCents),
		Currency:      stripe.String(chargeReq.Currency),
		Customer:      stripe.String(chargeReq.CustomerID),
		PaymentMethod: stripe.String(chargeReq.PaymentMethodID), //off-session payment method e.g saved card

		//Critical Off session Logic .. It will implement 3ds and other authentication automatically
		// confirm=true: Try to charge immediately (don't just create a draft).
		// off_session=true: Tell the bank "User is asleep, trust us".
		Confirm:     stripe.Bool(true),
		OffSession:  stripe.Bool(true),
		Description: stripe.String(chargeReq.Description),
	}
	// 3. Idempotency (Distributed System Safety)
	// Prevents double-charging if the network fails but Stripe succeeded.
	if chargeReq.ReferenceID != "" {
		params.IdempotencyKey = stripe.String(chargeReq.ReferenceID) // Unique key for this payment attempt

	}
	//Metadata Optimization for better tracking
	//we check length to avoid allocationg a map if no metadata is provided
	//means if len is 0 we skip this step. metadata map means key-value pairs.
	// it helps in searching and filtering payments in the stripe dashboard.
	// e.g tenantID,invoiceID etc.
	if len(chargeReq.MetaData) > 0 {
		params.Metadata = make(map[string]string, len(chargeReq.MetaData))
		for k, v := range chargeReq.MetaData {
			params.Metadata[k] = v
		}
	}
	//Context Propagation
	//If the server shuts down or the request times out, this cancels then http request to Stripe.

	params.Context = ctx
	//Execute (Network Call)
	// 6. Execute (Network Call)
	// We use s.client.PaymentIntents, NOT paymentintent.New (which uses global state).
	//PaymentIntent means a charge attempt in stripe terminology. it may require multiple steps to complete
	pi, err := sg.client.PaymentIntents.New(params) //this lines means we are creating a payment intent in stripe
	// 7. Error Translation
	if err != nil {
		return nil, sg.mapStripeError(err)
	}

	// 8. Business Logic Verification
	// Network success != Payment success. We must check the Status.
	if pi.Status != stripe.PaymentIntentStatusSucceeded {
		// e.g., "requires_action" means the bank demanded 3D Secure (OTP).
		// Since this is off-session, we can't show a popup, so we fail.
		return &PaymentResult{
			TransactionID: pi.ID,
			status:        PaymentStatus(pi.Status),
			RawResponse:   string(pi.LastResponse.RawJSON),
			PaidAt:        time.Now(),
		}, fmt.Errorf("%w: status is %s (requires user action)", ErrPaymentFailed, pi.Status)

	}

	// 9. Success
	return &PaymentResult{
		TransactionID: pi.ID,
		status:        PaymentStatus(pi.Status),
		RawResponse:   string(pi.LastResponse.RawJSON),
		PaidAt:        time.Now(),
	}, nil
}

// mapStripeError converts external library errors into Domain Errors.
// This prevents 'stripe-go' imports from leaking into our Business Service layer.
func (sg *StripeGateway) mapStripeError(err error) error {
	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) {
		switch stripeErr.Code {
		case stripe.ErrorCodeCardDeclined:
			return fmt.Errorf("%w: card was declined (%s)", ErrPaymentFailed, stripeErr.Msg)
		case stripe.ErrorCodeExpiredCard:
			return fmt.Errorf("%w: card has expired", ErrPaymentFailed)
		case stripe.ErrorCodeBalanceInsufficient:
			return fmt.Errorf("%w: insufficient funds", ErrPaymentFailed)
		case stripe.ErrorCodeIdempotencyKeyInUse:
			return fmt.Errorf("system conflict: idempotency key collision (check reference ID)")
		}

		// Check HTTP status for outages
		if stripeErr.HTTPStatusCode >= http.StatusInternalServerError {
			return ErrProviderDown
		}
	}
	return fmt.Errorf("gateway internal error: %w", err)
}
func (sg *StripeGateway) GetPaymentStatus(ctx context.Context, id string) (PaymentStatus, error) {
	//call stripe API to get payment intent status
	pi, err := sg.client.PaymentIntents.Get(id, nil) // it will fetch payment intent by id from stripe
	if err != nil {
		return PaymentStatusUnknown, sg.mapStripeError(err)
	}
	// Map Stripe Status to Domain Status
	switch pi.Status {

	case "succeeded":
		return PaymentSucceeded, nil
	case "requires_payment_method", "canceled":
		return PaymentFailed, nil
	case "processing":
		return PaymentStatusPending, nil
	default:
		// "requires_action", "requires_capture", etc. treat as pending or specialized status
		return PaymentStatusPending, nil
	}

}
