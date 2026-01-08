// services/billing-service/internal/payment/models.payment.go.go
package payment

import (
	"context"

	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/accounts"
	"github.com/google/uuid"
)

// WE define an AccountProvider interface which will be used to interact with
// different payment gateways like Stripe, PayPal etc.
//AccountProvider interface allows payment service to  to ask "Give me he billing details for this tenant ID"

// PaymentGateway interface abstract the actual  Money Mover (payment processor like Stripe etc)
// It accepts Context for cancellation or timeouts Propagation
type PaymentGateway interface {

	//ChargeAttempt tries to charge the given amount (in cents) to the payment method on file for the specified tenant.
	//it executes a synchronous ,off-session charge..Means there is no user interaction involved
	ChargeAttempt(ctx context.Context, paymentReq PaymentRequest) (*PaymentResult, error)
}

// AccountProvider interface  abstracts the source of truth for user billing account information. Means it provides methods to retrieve billing account details for a given tenant.
// This bridges the gap between the Billing service and Auth/Account service.
type AccountProvider interface {
	// GetBillingDetails fetches the Stripe ID and Payment Method for a tenant.
	// It returns the domain 'Account' struct we defined in internal/accounts.
	GetBillingAccountDetails(ctx context.Context, tenantID uuid.UUID) (*accounts.Account, error)
}

// InvoiceUpdater abstracts the side-effects of a successful payment.
// We prefer small interfaces over large ones (Interface Segregation Principle).
type InvoiceUpdater interface {
	//transactionID is the payment gateway transaction reference
	MarkPaid(ctx context.Context, invoiceID uuid.UUID, transactionID string) error
}
