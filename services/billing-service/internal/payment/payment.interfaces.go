// services/billing-service/internal/payment/payment.interfaces.go
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

// AccountProvider lets Payment Service fetch data without knowing about the DB.
// it is exactly same as accounts.AccountStore
// The Store layer will implement this.
type AccountProvider interface {
	GetBillingAccountDetails(ctx context.Context, tenantID uuid.UUID) (*accounts.Account, error)
}

// InvoiceUpdater interface defines methods to update invoice payment status
type InvoiceUpdater interface {
	MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID, transactionID string) error
}
