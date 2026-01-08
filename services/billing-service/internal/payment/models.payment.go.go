package payment

import (
	"errors"

	"github.com/google/uuid"
)

// These structuirs are data transfer object of payment service
// StandardPayment Errors
var (
	ErrPaymentFailed   = errors.New("payment Gateway rejected the transaction")
	ErrInvalidAmount   = errors.New("invalid payment amount")
	ErrNoPaymentMethod = errors.New("customerr has no valid payment method on file")
	ErrAlreadyPaid     = errors.New("invoice is already paid")
	ErrProviderDown    = errors.New("payment provider is currently unavailable") //e.g stripe API down

)

//PaymentRequest Encapsulates all data needed to a payment transaction

type PaymentRequest struct {
	ReferenceID     uuid.UUID         // Unique identifier for the payment transaction
	AmountCents     int64             // Amount to be charged in cents
	Currency        string            // Currency code (e.g., "USD")
	CustomerID      string            // Identifier for the customer in the payment gateway
	PaymentMethodID string            // Identifier for the payment method to be used e.g stripe payment method id..The token
	Description     string            // Appears on BANK/CC statements
	MetaData        map[string]string // Additional metadata for the transaction context tags (tenantID,invoiceID etc) ..it helps in searching and filtering payments in the payment gateway dashboard

}

//PaymentResult Represents the outcome of a payment transaction

type PaymentResult struct {
	TransactionID string // Unique identifier for the payment transaction in the payment gateway (e.g., Stripe Charge ID)
	status        string // Status of the payment (e.g., "succeeded", "failed")
	RawResponse   string // Raw response from the payment gateway for logging/debugging purposes
	PaidAt        int64  // Timestamp when the payment was completed
}
