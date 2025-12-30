// services/billing-service/internal/billingTypes/types.billing.go
package billingtypes

type UsageType string

const (
	ShipmentCreated UsageType = "SHIPMENT_CREATED"
	APIRequest      UsageType = "API_REQUEST"
)

// ... existing code ...

type TransactionType string

const (
	// DEBIT: The customer owes us money (Positive amount in Ledger)
	TransactionTypeDebit TransactionType = "DEBIT"

	// CREDIT: We owe the customer money (Negative amount in Ledger)
	TransactionTypeCredit TransactionType = "CREDIT"
)
