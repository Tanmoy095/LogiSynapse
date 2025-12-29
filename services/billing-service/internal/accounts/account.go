// services/billing-service/internal/accounts/account.go
package accounts

import "github.com/docker/distribution/uuid"

// PlanType represents the subscription plan type for a billing account
type PlanType string

const (
	FreePlan       PlanType = "FREE"
	ProPlan        PlanType = "PRO"
	EnterprisePlan PlanType = "ENTERPRISE"
)

//Account represents a billing entity for a User/Tanant/Organization

type Account struct {
	ID              uuid.UUID // Unique identifier for the billing account
	Email           string    // Contact email for billing communications forr notification or invoices
	CurrentPlan     PlanType  // Current subscription plan of the account}
	PaymentMethodID string    // Reference to the payment method on file (e.g., Stripe Payment Method ID)
}

//SubscriptionStatus tracks the status of subscriptions

type SubscriptionStatus string

const (
	Active   SubscriptionStatus = "ACTIVE"
	PastDue  SubscriptionStatus = "PAST_DUE"
	Canceled SubscriptionStatus = "CANCELED"
)

// Subscription represents a billing subscription for an account

//zone 1:for the first  10 miles the price is $1 per mile
//zone 2:for any distance after 10 miles the price is $0.75 per mile
//How does the driver calculate the bill? They don't charge $0.50 for the whole trip just because you went far. They split the trip into "legs."

//Leg 1 (zone 1): 10 miles at $1/mile = $10
//Leg 2 (zone 2): 5 miles at $0.75/mile = $3.75
//Total fare: $10 + $3.75 = $13.75
