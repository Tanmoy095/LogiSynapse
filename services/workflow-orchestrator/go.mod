module github.com/Tanmoy095/LogiSynapse/services/workflow-orchestrator

go 1.24.4

replace github.com/Tanmoy095/LogiSynapse/shared => ../../shared

replace github.com/Tanmoy095/LogiSynapse/services/shipment-service => ../shipment-service

require go.temporal.io/sdk v1.38.0

require github.com/lib/pq v1.10.9 // indirect

require (
	github.com/Tanmoy095/LogiSynapse/services/shipment-service v0.0.0-20251126141832-b622e1cd448d
	github.com/Tanmoy095/LogiSynapse/shared v0.0.0-20251126141832-b622e1cd448d
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/nexus-rpc/sdk-go v0.5.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/segmentio/kafka-go v0.4.49 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.temporal.io/api v1.54.0 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/grpc v1.77.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
