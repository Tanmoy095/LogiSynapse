# Shipment Service — Design, current implementation & developer guide

This document explains how the Shipment Service behaves end-to-end, how requests from the GraphQL gateway are handled by the gRPC server and service layers, what the Postgres store does, the external dependencies (Shippo, Kafka), where we are now with the recent enum changes, and recommended project structure / best practices for `internal` vs `pkg` packages.

Contents

- Overview & actors
- End-to-end: CreateShipment and GetShipments flows (call sequence + file mapping)
- Files and folders — purpose and where to look
- Current state (what I changed / where we are now)
- DB mapping & migrations notes
- Testing & local dev
- Best practice: internal vs pkg and suggested refactor

## Overview & actors

Actors:

- GraphQL Gateway — frontend gateway that exposes GraphQL to clients and acts as a gRPC client to the Shipment Service.
- Shipment gRPC Server — `services/shipment-service/handler/grpc/shipment.handler.grpc.go` implements the gRPC API defined in `shared/proto/shipment.proto`.
- Service layer — `services/shipment-service/service/shipment.service.go` contains business logic (validation, calling Shippo, publishing Kafka events).
- Store — `services/shipment-service/store/postgres.go` is the Postgres-backed persistence layer implementing `store.ShipmentStore`.
- External services — Shippo (HTTP API for creating shipments and getting rates), Kafka (event bus).

## End-to-end flows

1. CreateShipment (GraphQL -> gRPC -> Service -> Shippo -> Store -> Kafka)

- GraphQL Gateway: receives GraphQL mutation `createShipment(input)`.

  - Calls gRPC CreateShipment on the ShipmentService (generated client at `services/graphql-gateway/client/shipment.client.go`).

- gRPC handler: `handler/grpc/shipment.handler.grpc.go`

  - `CreateShipment(ctx, req *proto.CreateShipmentRequest)`
  - Converts `req` (proto) to `models.Shipment` via `toModelShipment(req)`.
  - Calls `svc.CreateShipment(ctx, shipment)`.

- Service: `service/shipment.service.go`

  - Validates required fields and package dimensions.
  - Builds a Shippo API request (parcels with dimensions) and posts to Shippo.
  - Decodes Shippo response (object_id, tracking_number, tracking_url, carrier, status).
  - Maps Shippo status (string) to `proto.ShipmentStatus` using `mapShippoStatusToProto`.
  - Fills the internal `models.Shipment` with Shippo data.
  - Calls `store.CreateShipment(ctx, shipment)`.
  - Publishes `shipment.created` event to Kafka (if producer configured).
  - Returns the persisted shipment.

- Store: `store/postgres.go`
  - Persists the shipment into `shipments` table.
  - Reads/writes `status` as TEXT in DB, but service and models use `proto.ShipmentStatus`.
  - To bridge this, read/write convert enum <-> string (helper `parseStatusStringToProto` and using `status.String()` when writing).

2. GetShipments (GraphQL -> gRPC -> Service -> Store)

- Gateway calls gRPC `GetShipments` with filters (origin, status enum, destination, limit, offset).
- gRPC handler forwards the request to service: `s.service.GetShipments(req.Origin, req.Status, req.Destination, req.Limit, req.Offset)`.
- Service converts status (if needed) and calls `store.GetShipments(...)` which queries DB and returns a slice of `models.Shipment`.
- Handler converts each `models.Shipment` to `proto.Shipment` and returns to gateway.

## Files and folders — purpose and where to look

- services/shipment-service/
  - `cmd/main.go` — start server, create store and service, register gRPC server.
  - `config/config.go` — env parsing and DB url builder.
  - `handler/grpc/shipment.handler.grpc.go` — gRPC handlers and helpers to convert between proto and internal models.
  - `service/shipment.service.go` — business logic.
  - `store/postgres.go` — Postgres implementation.
  - `internal/models/shipment.models.go` — domain model shared between service and handler.
  - `internal/kafka/producer.go` — Kafka producer wrapper.
  - `db/migrations/*` — SQL migrations.

## Current state (what I changed / where we are now)

- Proto: `shared/proto/shipment.proto` now includes `PRE_TRANSIT` and `CANCELLED` enum values.
- Models: `internal/models/shipment.models.go` uses `proto.ShipmentStatus` for `Status` and exposes fields used by service and store (tracking/dimensions).
- Service: uses `mapShippoStatusToProto` to convert Shippo text statuses into proto enums; uses proto enum constants for comparisons.
- Store: converts proto enum -> string when writing, and string -> proto enum when reading. Helpers were added: `parseStatusStringToProto`.
- Build: `go build` for `services/shipment-service` completes successfully after these changes.

## DB mapping & migrations notes

- Current migration creates `status TEXT`. Two possible approaches:
  1. Keep `status TEXT` and always map enum<->string when reading/writing. This is flexible and human readable in DB.
  2. Create a Postgres `ENUM` type for `shipment_status` and use it in the table. This enforces allowed values at DB level but requires migration steps when enums change.

Recommendation: Start with TEXT + application-level mapping during development. For production, migrate to Postgres ENUM or a lookup table once the statuses are stable.

## Testing & local dev

- Manual run: `cd services/shipment-service && go run ./cmd` (ensure env vars set)
- Unit tests: add tests under `service` and `store`, mock `http.Client` for Shippo and `sql.DB` or use a test container for Postgres.

## Best practice: `internal` vs `pkg` and suggested refactor

Current situation: The repo has multiple `internal` packages (one under `services/shipment-service/internal` and another under `services/graphql-gateway/internal`). In Go, `internal` packages prevent import from outside the tree which is good for encapsulation.

However, some code (shared types or utilities) may be used across services (e.g., models or helpers). For cross-service reusable code, create a `pkg/` top-level module or use the existing `shared/` module.

Suggested structure:

- Keep service-specific code in `services/<service>/internal/...` (private to that service).
- Move cross-service, versioned, reusable code into `pkg/` or `shared/` packages:
  - `shared/proto` (already exists) — keep proto files here.
  - `pkg/logging` — logging helpers if needed.
  - `pkg/config` — shared configuration parsing if multiple services share a pattern.

Why `pkg`?

- `internal` is private to the containing module path. Use it to prevent accidental imports.
- `pkg` is the conventional place for libraries intended to be imported by other services.

Concrete suggestion for this repo:

1. Keep `services/*/internal` for private implementation details.
2. Keep `shared/proto` as the canonical place for protobuf definitions.
3. Create `pkg/` for cross-cutting Go packages you want other services to import (e.g., `pkg/errs`, `pkg/httpclient`, `pkg/kafka` wrappers, `pkg/models` if you choose to centralize domain types beyond proto).

Example layout:

```
/
├─ pkg/
│  ├─ kafka/         # thin wrapper for producer + testing helpers
│  ├─ db/            # common DB helpers or test fixtures
│  └─ config/        # shared config parsing if needed
├─ shared/
│  └─ proto/         # proto files and generated Go
├─ services/
│  ├─ shipment-service/
│  │  └─ internal/   # private to shipment-service
│  └─ graphql-gateway/
│     └─ internal/   # private to gateway
```

Migration notes:

- If you create `pkg/models` and move domain structs there, ensure you do not re-introduce cyclic imports. Prefer small focused packages (kafka client, config). Domain models that must stay private should remain in `internal/models`.

## Conclusion and immediate next steps

1. Keep the `internal` packages as-is for service-private code.
2. Move cross-service reusable helpers to `pkg/` (start with kafka wrapper and config if you need cross-service sharing).
3. Add unit tests for the service and store (I can scaffold test files if you want).

If you'd like, I can:

- Create a `pkg/kafka` wrapper and move `internal/kafka/producer.go` there (or create a thin adapter) so the gateway or other services can reuse it.
- Scaffold unit tests for `mapShippoStatusToProto` and DB mapping helpers.

Which of those should I do next?
