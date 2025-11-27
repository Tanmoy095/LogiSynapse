package config

import (
	"os"
	// Alias the import because the package name 'config' conflicts with this package name
	sharedConfig "github.com/Tanmoy095/LogiSynapse/shared/config"
)

type ShipmentConfig struct {
	*sharedConfig.CommonConfig        // Embed the shared struct
	ShippoKey                  string // Specific to this service only!
}

// Rename 'Load' to 'LoadConfig' so it matches your main.go call
func LoadConfig() *ShipmentConfig {
	return &ShipmentConfig{
		// Use the new function name from Shared
		CommonConfig: sharedConfig.LoadCommonConfig(),
		ShippoKey:    os.Getenv("SHIPPO_API_KEY"),
	}
}
