package config

import (
	"fmt"
	"os"

	"github.com/Tanmoy095/LogiSynapse/shared/config"
)

type BillingConfig struct {
	CommonConfig *config.CommonConfig // this helps to access DB and RabbitMQ configs directly
	//Domain-specific configs can be added here in future if needed
	StripeSecretKey string // Stripe API secret key
}

// LoadConfig loads the billing service configuration
func LoadConfig() (*BillingConfig, error) {
	//Load shared configuration
	common := config.LoadCommonConfig()
	//Load billing-specific configuration
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		//sent error messege stripe key needed and default set locally
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is required")

	}
	return &BillingConfig{
		CommonConfig:    common,
		StripeSecretKey: stripeKey,
	}, nil

}
