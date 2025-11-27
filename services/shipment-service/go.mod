module github.com/Tanmoy095/LogiSynapse/services/shipment-service

go 1.24.4

replace github.com/Tanmoy095/LogiSynapse/shared => ../../shared

require (
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.77.0
)

require (
	github.com/Tanmoy095/LogiSynapse/shared v0.0.0-20251126141832-b622e1cd448d
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.49 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
