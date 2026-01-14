// services/billing-service/internal/payment/models.payment.go.go
package payment

import (
	"errors"
	"time"

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
	ReferenceID     string            // Unique identifier for the payment transaction
	AmountCents     int64             // Amount to be charged in cents
	Currency        string            // Currency code (e.g., "USD")
	CustomerID      string            // Identifier for the customer in the payment gateway
	PaymentMethodID string            // Identifier for the payment method to be used e.g stripe payment method id..The token
	Description     string            // Appears on BANK/CC statements
	MetaData        map[string]string // Additional metadata for the transaction context tags (tenantID,invoiceID etc) ..it helps in searching and filtering payments in the payment gateway dashboard

}

//PaymentResult Represents the outcome of a payment transaction

type PaymentResult struct {
	TransactionID string        // Unique identifier for the payment transaction in the payment gateway (e.g., Stripe Charge ID)
	status        PaymentStatus // Status of the payment (e.g., "succeeded", "failed")
	RawResponse   string        // Raw response from the payment gateway for logging/debugging purposes
	PaidAt        time.Time     // Timestamp when the payment was completed
}

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentSucceeded     PaymentStatus = "SUCCEEDED"
	PaymentFailed        PaymentStatus = "FAILED"
	StatusRequiresAction PaymentStatus = "REQUIRES_ACTION"
	PaymentStatusUnknown PaymentStatus = "UNKNOWN"
)

// PaymentAttempt represents a record of a payment attempt in the system.
type PaymentAttempt struct {
	AttemptID         uuid.UUID // Unique identifier for the payment attempt
	InvoiceID         uuid.UUID // Associated invoice ID. it helps in linking payment attempt to specific invoice
	TenantID          uuid.UUID // Tenant identifier for multi-tenant systems
	Provider          string    // Payment provider used (e.g., "Stripe")
	ProviderPaymentID string    // Identifier from the payment provider (e.g., Stripe PaymentIntent ID)
	Status            PaymentStatus
	AmountCents       int64
	Currency          string
	ErrorCode         *string // Pointer to allow NULL
	ErrorMessage      *string // Pointer to allow NULL
	RetryCount        int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
