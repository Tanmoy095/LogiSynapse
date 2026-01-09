# Phase 3 — Master Implementation Plan (Billing & Payment Integration)

This document outlines a detailed, actionable plan to add a secure billing capability to LogiSynapse. It includes service design, database schema and migration, proto and API design, Temporal workflow changes, integration points with existing services, testing, local/dev tooling, Kubernetes guidance, CI/CD, and operational notes.

Purpose: implement a `billing-service` that creates Stripe PaymentIntents, persists invoices, verifies Stripe webhooks, and signals Temporal workflows so shipments move from payment-pending to PRE_TRANSIT.

Audience: engineers working on the LogiSynapse monorepo. All steps are actionable and mapped to repository locations.

—

## Overview (what will change)

- Add `services/billing-service` (gRPC + small HTTP webhook server)
- Add DB migration to create `invoices` table in the app Postgres
- Add `billing.proto` (CreateInvoice, GetInvoice) and generate Go stubs
- Implement Stripe client wrapper and webhook verification
- Modify Temporal `CreateShipmentWorkflow` to call billing and wait for PaymentConfirmed signal
- Update `shipment-service` statuses and default behavior (PENDING_PAYMENT)
- Add GraphQL schema & resolver for fetching invoice client secret
- Update `communications-service` to send payment receipts and cancellation emails
- Deliver tests (unit, integration, E2E) and CI steps

—

## Part 1 — New Service: `billing-service`

Design goals

- Securely handle money-related workloads (never store raw card data)
- Be idempotent for webhook handling
- Minimal surface area: gRPC for internal calls, HTTP for Stripe webhook
- Keep online path low-latency: CreateInvoice should return Stripe client_secret quickly

  1.1 Repository layout

services/billing-service/

- cmd/main.go # gRPC server + webhook HTTP server (8081)
- proto/billing.proto # Billing protobuf (also mirrored to shared/proto if preferred)
- internal/
  - server/
    - billing_server.go # gRPC handlers
  - stripe/
    - client.go # Stripe wrapper: CreatePaymentIntent, VerifyWebhook
  - store/
    - postgres.go # Invoice persistence
  - webhook/
    - handler.go # HTTP POST /webhook
- db/migrations/
  - 00001_create_invoices_table.sql
- Dockerfile
- go.mod

  1.2 Database migration (`00001_create_invoices_table.sql`)

CREATE TABLE invoices (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
user_id TEXT NOT NULL,
shipment_id UUID NULL,
workflow_id TEXT NULL,
stripe_intent_id TEXT NOT NULL,
amount_cents BIGINT NOT NULL,
status TEXT NOT NULL, -- PENDING, PAID, FAILED
created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
CREATE INDEX idx_invoices_shipment_id ON invoices(shipment_id);
CREATE INDEX idx_invoices_workflow_id ON invoices(workflow_id);

Down script: DROP TABLE IF EXISTS invoices;

1.3 Proto: `billing.proto`

syntax = "proto3";
package billing;
option go_package = "github.com/Tanmoy095/LogiSynapse/services/billing-service/proto";

service BillingService {
rpc CreateInvoice(CreateInvoiceRequest) returns (CreateInvoiceResponse);
rpc GetInvoice(GetInvoiceRequest) returns (GetInvoiceResponse);
}

message CreateInvoiceRequest {
string shipment_id = 1;
string user_id = 2;
string workflow_id = 3; // Temporal workflow run id or business id
int64 amount_cents = 4;
}

message CreateInvoiceResponse {
string stripe_client_secret = 1;
string invoice_id = 2;
}

message GetInvoiceRequest { string shipment_id = 1; }
message GetInvoiceResponse { string status = 1; string stripe_client_secret = 2; }

Implementation note: keep the proto under `services/billing-service/proto/` or under `shared/proto/` if you want cross-service import.

1.4 Stripe client wrapper (internal/stripe/client.go)

- Read `STRIPE_SECRET_KEY` from env
- Expose: CreatePaymentIntent(ctx, amountCents, currency, metadata) -> (clientSecret, intentID, err)
- Expose: VerifyWebhookSignature(rawBody, sigHeader, webhookSecret) -> event
- Use `stripe-go` official SDK; in tests, allow a mock implementation.

Security: set `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET` as secrets (env, or Kubernetes secret manager). Log only non-sensitive metadata.

1.5 Billing gRPC flow — CreateInvoice

1. Validate request: amount > 0, user_id present
2. Call Stripe: CreatePaymentIntent with metadata map{"workflow_id": workflow_id, "shipment_id": shipment_id}
3. Persist invoice row (stripe_intent_id, PENDING)
4. Return `stripe_client_secret` and `invoice_id` to caller

Idempotency: if same workflow_id+shipment_id exists, return existing invoice and client_secret instead of creating duplicate intents.

1.6 HTTP Webhook server

- Bind to port `8081`
- Endpoint `POST /webhook`
- Verify signature using `STRIPE_WEBHOOK_SECRET`
- For event `payment_intent.succeeded`:
  - parse `payment_intent` object
  - get `intent.Metadata["workflow_id"]` (and shipment_id)
  - update invoice status to `PAID`
  - call Temporal client (see below) to signal workflow (signal name: "PaymentConfirmed", payload: invoice.id / intent id)
  - publish event to Kafka (optional) or send an internal event to communications-service

Webhook idempotency: Stripe may deliver events multiple times; check invoice status before double-updating and make signals idempotent (Temporal signals should tolerate duplicates).

—

## Part 2 — Update Existing Services

2.1 workflow-orchestrator (changes)

- Update `CreateShimentWorkflow` to include a billing step and wait for signal:
  1. ACTIVITY_CallShippoAPI (optional: create shipment draft in Shippo)
  2. ACTIVITY_SaveShipmentToDB (create DB record with status=PENDING_PAYMENT)
  3. ACTIVITY_CallBillingService (call billing.CreateInvoice) — this is an activity that does short gRPC call
  4. Wait for Temporal Signal `PaymentConfirmed` (workflow.Signal channel or Workflow.SignalWithStart patterns)
     - Use `workflow.NewChannel(ctx)` and `Select` with `Receive` + `workflow.NewTimer(ctx, 24*time.Hour)` to implement 24h timeout
  5. If `PaymentConfirmed` received: ACTIVITY_UpdateShipmentToPreTransit
  6. If timeout: ACTIVITY_CancelShipment (void label via Shippo, mark shipment CANCELLED)

Implementation tips:

- Use a deterministic signal name `PaymentConfirmed` and pass the `workflow_id` in CreateInvoice metadata to correlate
- Use robust retries and idempotent activities

  2.2 shipment-service changes

- Update `shared/proto/shipment.proto` to include `PENDING_PAYMENT` in the `ShipmentStatus` enum and `PRE_TRANSIT` if not present (you already have PRE_TRANSIT; ensure PENDING_PAYMENT present).
- Update `services/shipment-service/service/shipment.service.go` to set `status = PENDING_PAYMENT` when creating shipments (or when workflow reaches billing step)
- Ensure `UpdateShipment` and any status transitions are validated (only allow PRE_TRANSIT -> other states as appropriate)

  2.3 graphql-gateway changes

- Add to `graph/schema/schema.graphqls`:
  ```graphql
  type Invoice {
    id: ID!
    status: String!
    stripeClientSecret: String
  }
  extend type Query {
    getInvoice(shipmentId: ID!): Invoice
  }
  ```
- Add gRPC client `services/graphql-gateway/client/billing.client.go` similar to `shipment.client.go` and call `billing-service`'s `GetInvoice`
- Resolver `GetInvoice` returns status and stripeClientSecret

  2.4 communications-service updates

- Subscribe to `shipment.status.updated` events (or read from Kafka topic)
- When status==PRE_TRANSIT -> send payment receipt & confirmation email
- When status==CANCELLED (due to payment timeout) -> send cancellation email

—

## Part 3 — Execution & Testing Sequence

Local/dev prerequisites

- Postgres (app + temporal-db)
- Temporal server & UI
- RabbitMQ
- Kafka (optional for event routing; communications-service uses Kafka consumer)
- Stripe CLI (`stripe`) to emulate events and trigger `payment_intent.succeeded`

  3.1 Docker Compose updates

- Add `billing-service` to `docker-compose.yml` with:

  - ports: 8081:8081 (webhook), grpc port if needed (e.g., 50052)
  - env_file: .env (contains STRIPE keys)
  - depends_on: postgres, temporal, rabbitmq

  3.2 Manual dev flow (Happy path)

1. Start infra: `docker-compose up --build postgres temporal temporal-ui rabbitmq kafka billing-service workflow-orchestrator shipment-service graphql-gateway communications-service`
2. Create shipment via GraphQL `createShipment` mutation:
   - Expected: DB shipment row created with status `PENDING_PAYMENT`
   - Workflow visible in Temporal UI and waiting for signal
3. Use Stripe CLI to emulate payment success:
   - `stripe listen --forward-to localhost:8081/webhook`
   - `stripe trigger payment_intent.succeeded` (or use `stripe trigger` with appropriate payload)
4. Billing-service webhook receives event, verifies signature, updates invoice status to `PAID`, signals Temporal workflow
5. Workflow continues, calls ACTIVITY_UpdateShipmentToPreTransit; shipment status becomes `PRE_TRANSIT`
6. Communications-service receives status update and sends receipt

3.3 Timeout test (Abandonment)

1. Create shipment -> workflow waits for payment
2. Do not trigger Stripe; wait for 24h or accelerate using Temporal test time advance
3. Workflow timer fires -> ACTIVITY_CancelShipment executes -> shipment.status becomes `CANCELLED`
4. Communications-service sends cancellation email

—

## Testing Strategy (unit -> integration -> E2E)

Unit tests

- stripe client: mock Stripe API by abstracting SDK behind interface and injecting mock for CreatePaymentIntent/VerifyWebhook
- DB store: use in-memory or mocked DB interface
- activities: unit-test activity functions using injected mocks

Integration tests

- Start Postgres and run migrations (Testcontainers recommended)
- Start local Temporal test server or use Temporal test suite to drive workflow state machine
- Use Stripe test keys and the Stripe CLI for webhook delivery

E2E (CI job / manual)

- Bring up infra in ephemeral environment (docker-compose or Kind cluster)
- Run full scenario: createShipment -> stripe CLI -> verify timelines, DB state, Temporal workflow completion

Automated test suggestions

- Add a GitHub Actions workflow `ci.yml` that runs `go test ./...` and specific integration suites
- Add an `integration` stage that can be toggled by environment variable or by running in a dedicated runner with Docker socket

—

## Concurrency, Idempotency & Best Practices

Idempotency

- CreateInvoice must be idempotent keyed by `workflow_id` or (workflow_id, shipment_id)
- Webhooks must check invoice status before performing updates; guard against double-processing

Temporal signal idempotency

- Use signal payload with invoice id and idempotent handling inside workflow (e.g., set flag or check DB)

Database transactions

- Persist invoice after creating PaymentIntent; prefer storing PaymentIntent ID from Stripe so webhook can reconcile

Retries and backoff

- Use exponential backoff for external calls (Stripe network calls). Activities run under Temporal retry policies.

Observability and tracing

- Add structured logs (slog/zerolog) and include correlation ids: `workflow_id`, `shipment_id`, `invoice_id`, `stripe_intent_id`
- Add OpenTelemetry tracing for gRPC, HTTP, and DB calls; propagate context through activities

Secrets

- Use Kubernetes secrets or Vault in production; locally `.env` but do not commit sensitive keys

—

## Kubernetes & Production Considerations

Containerization

- Build `billing-service` image and push to registry

Kubernetes manifests (recommended minimal set)

- `deployment.yaml` for `billing-service` (replicaCount 2) with liveness/readiness probes
- `service.yaml` expose gRPC and webhook HTTP (ClusterIP)
- `ingress.yaml` for public webhook endpoint with TLS (or use a private tunnel for Stripe and use internal service with externalName)
- `secret.yaml` for STRIPE_SECRET_KEY and STRIPE_WEBHOOK_SECRET
- `horizontalpodautoscaler.yaml` based on CPU and concurrency metrics

Webhook exposure

- Stripe requires externally reachable webhook URL. For dev: use `stripe listen` to forward to localhost. For prod: use ingress with TLS + public domain

Database

- Ensure migration job runs at deploy (Job or init container) with locking mechanism (like `golang-migrate` or `goose`) and that migration is safe for multiple replicas

Scaling & concurrency

- Billing service is stateless; scale horizontally
- Avoid sticky sessions; use DB for persistence

Monitoring & alerts

- Prometheus metrics (request rate, error rate, invoice creation rate, webhook verification failures)
- Alert on failed webhook deliveries and high ratio of FAILED invoices

—

## CI/CD (GitHub Actions example)

Pipeline stages

1. `lint` — `golangci-lint`
2. `unit-tests` — `go test ./...`
3. `build` — `go build ./...` and `docker build` for `billing-service`
4. `integration` (optional) — start ephemeral infra (docker-compose), run a small e2e script that uses Stripe CLI
5. `push` — push images to registry
6. `deploy` — apply manifests to staging cluster

Security in CI

- Run `go vet`, `gosec` or equivalent
- Scan dependencies (Dependabot or `govulncheck`)

—

## Delivery Checklist

- [ ] `services/billing-service` repository skeleton
- [ ] DB migration added and tested locally
- [ ] `billing.proto` added and code generated
- [ ] Stripe client wrapper implemented (with mocks)
- [ ] gRPC CreateInvoice and GetInvoice implemented
- [ ] Webhook HTTP server implemented and tested with Stripe CLI
- [ ] Temporal workflow updated and working in Temporal UI
- [ ] GraphQL gateway schema and resolver updated
- [ ] communications-service notifications implemented
- [ ] Unit tests and integration tests added
- [ ] Docker Compose updated and manual happy-path validated
- [ ] Kubernetes manifests created and deployment tested
- [ ] CI pipeline updated and passing

—

## Notes & Risks

- Be careful to never log or persist card details. Only store Stripe IDs and metadata.
- Stripe webhooks require signature verification — misconfiguration may cause missed events.
- Temporal timers and test time manipulation will speed up timeout testing (do not reduce production timeout).
- For early development, Stripe test mode + `stripe listen` is the fastest path.

—

## Recommended Timeline (Rough)

- Week 1: Scaffold `billing-service`, add migration, proto, and basic CreateInvoice flow (no webhooks)
- Week 2: Implement webhook verification, signal Temporal, update workflow and shipment-state transitions
- Week 3: Add GraphQL integration, communications-service updates, and end-to-end testing with Stripe CLI
- Week 4: Harden (logging, retries, metrics), add CI integration and Kubernetes manifests

—

## Contact Points (owners)

- Billing service implementation: backend engineer
- Temporal workflow updates: workflow engineer
- GraphQL and frontend contract: API engineer
- Communications & notifications: notifications engineer
- DevOps & infra: platform engineer

—

Appendix: Helpful commands

Start Stripe listener (dev):

```bash
stripe listen --forward-to localhost:8081/webhook
```

Run migrations locally (example):

```bash
cd services/billing-service/db/migrations
goose postgres "postgres://postgres:postgres@localhost:5432/shipments?sslmode=disable" up
```

Call CreateInvoice (example gRPC flow):

1. From gateway or service: call gRPC `CreateInvoice` with workflow_id
2. The response returns `stripe_client_secret` used by frontend to confirm card

—

End of Phase 3 Master Implementation Plan
