module github.com/Tanmoy095/LogiSynapse/services/graphql-gateway

go 1.24.4

replace github.com/Tanmoy095/LogiSynapse/shared => ../../shared

require (
	github.com/99designs/gqlgen v0.17.84
	github.com/vektah/gqlparser/v2 v2.5.31
)

require (
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

require (
	github.com/Tanmoy095/LogiSynapse/shared v0.0.0-20251126141832-b622e1cd448d
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	google.golang.org/grpc v1.76.0
)
