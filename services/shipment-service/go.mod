module github.com/Tanmoy095/LogiSynapse/shipment-service

go 1.24.4

replace github.com/Tanmoy095/LogiSynapse/shared => ../../shared

replace github.com/Tanmoy095/LogiSynapse/pkg/kafka => ../../pkg/kafka

require (
	github.com/Tanmoy095/LogiSynapse/pkg/kafka v0.0.0-00010101000000-000000000000
	github.com/Tanmoy095/LogiSynapse/shared v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
	google.golang.org/grpc v1.76.0
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.49 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
