// services/billing-service/internal/billingTypes/types.billing.go
package billingtypes

type UsageType string

const (
	ShipmentCreated UsageType = "SHIPMENT_CREATED"
	APIRequest      UsageType = "API_REQUEST"
)
