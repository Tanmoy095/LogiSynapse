package pricing

import (
	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/google/uuid"
)

// priceRule represents a pricing rule for different usage tiers
type PriceRule struct {
	TenantID       *uuid.UUID // nil means default rule
	UsageType      billingtypes.UsageType
	UnitPriceCents int64  // price per unit in cents
	Currency       string // currency code, e.g., "USD"
}
