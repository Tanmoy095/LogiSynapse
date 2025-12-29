// services/billing-service/internal/pricing/store/pricing_store.go
package pricing_store

import (
	"context"
	"time"

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

type PricingStore interface {
	// GetPriceRules retrieves pricing rules for a given plan
	// GetPrice returns the price rule active at a specific point in time ('at').
	// Strategy: It looks for a tenant-specific price first. If none found, looks for default (tenant_id IS NULL).
	GetPriceRules(ctx context.Context, usageType billingtypes.UsageType, tenantID uuid.UUID, at time.Time) (PriceRule, error)
}
