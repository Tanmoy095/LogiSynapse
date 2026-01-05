// services/billing-service/internal/store/postgres/pricingRule_store.postgres.go

package Postgres_Store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/pricing"
	"github.com/google/uuid"
)

type PostgresPricingStore struct {
	db *sql.DB
}

func NewPostgresPricingStore(db *sql.DB) *PostgresPricingStore {
	return &PostgresPricingStore{db: db}
}

// Implement PricingStore methods here
// GetPrice returns the price rule active at a specific point in time ('at').
// Strategy: It looks for a tenant-specific price first. If none found, looks for default (tenant_id IS NULL).
// Tenant specific price means there is a row with tenant_id. tenant id IS NULL means default price for all tenants.

func (ps *PostgresPricingStore) GetPriceRules(ctx context.Context, usageType billingtypes.UsageType, tenantID uuid.UUID, at time.Time) (pricing.PriceRule, error) {
	// 	Implementation here
	// Query Strategy:
	// 1. Match Usage Type
	// 2. Match Tenant OR NULL (Default)
	// 3. Ensure 'at' is within [effective_from, effective_to]
	// 4. Order by tenant_id (Specific first), then date

	query := `
		SELECT tenant_id, unit_price_cents, currency
		FROM pricing_rules
		WHERE usage_type = $1  
		AND (tenant_id = $2 OR tenant_id IS NULL)
		AND effective_from <= $3
		AND (effective_to IS NULL OR effective_to > $3)
		ORDER BY 
		  tenant_id NULLS LAST, -- Prioritize specific tenant rules over NULL (default)
		  effective_from DESC   -- Get the most recent matching rule
		LIMIT 1;  --limit 1 means we want only one record
		
		`
	var rule pricing.PriceRule

	// We pass `usageType` as a query parameter to filter DB rows by usage type.
	// The SQL SELECT does not include `usage_type`, so we set it here so the
	// returned PriceRule contains the correct UsageType.
	rule.UsageType = usageType // because we are passing usage type in query to match usage type

	// Read tenant_id into a sql.NullString so we can detect NULL values (default rule).
	// The SELECT lists tenant_id first, so we scan into tenantIDStr first as well.
	// - If tenant_id is NULL in the DB, tenantIDStr.Valid will be false and we set
	//   rule.TenantID = nil to represent a default rule.
	// - If tenant_id is present, parse it into a uuid.UUID and set rule.TenantID
	//   to its pointer so the caller sees which tenant this rule belongs to.
	var tenantIDStr sql.NullString // to handle NULL tenant_id
	err := ps.db.QueryRowContext(ctx, query, usageType, tenantID, at).Scan(
		&tenantIDStr,
		&rule.UnitPriceCents,
		&rule.Currency,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rule, fmt.Errorf("no active pricing rule found for %s", usageType)
		}
		return rule, fmt.Errorf("database error looking up price: %w", err)
	}

	if tenantIDStr.Valid {
		// tenant_id was present in DB - parse the string value into a UUID.
		tid, err := uuid.Parse(tenantIDStr.String)
		if err != nil {
			// If DB contains malformed UUID, return an error rather than ignore it.
			return rule, fmt.Errorf("invalid tenant_id in db: %w", err)
		}
		rule.TenantID = &tid
	} else {
		// tenant_id was NULL in DB => this is a default rule for all tenants.
		rule.TenantID = nil
	}

	return rule, nil

}
