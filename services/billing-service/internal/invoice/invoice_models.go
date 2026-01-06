// services/billing-service/internal/invoice/invoice_models.go

package invoice

import (
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceDraft     InvoiceStatus = "DRAFT"
	InvoiceFinalized InvoiceStatus = "FINALIZED"
	InvoicePaid      InvoiceStatus = "PAID"
	InvoiceVoid      InvoiceStatus = "VOID"
)

type Invoice struct {
	InvoiceID  uuid.UUID
	TenantID   uuid.UUID
	Year       int
	Month      int
	TotalCents int64
	Currency   string // e.g., "USD"
	Status     InvoiceStatus
	Lines      []InvoiceLine
	CreatedAt  time.Time
}
type InvoiceLine struct {
	ID             uuid.UUID
	UsageType      billingtypes.UsageType
	Quantity       int64
	UnitPriceCents int64  // price per unit in cents
	LineTotalCents int64  // total price for this line in cents
	Description    string // e.g., "Shipment Created"
}
