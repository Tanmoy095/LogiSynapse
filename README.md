# LogiSynapse

> **Modern, event-driven shipment and billing platform with distributed workflow orchestration**

LogiSynapse is a production-grade microservices platform built with **Go**, **gRPC**, **GraphQL**, **Temporal.io**, and **PostgreSQL**. It demonstrates enterprise-grade patterns for logistics and billing operations, including workflow orchestration, event-driven architecture, and third-party API integration (Shippo, Stripe).

---

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Service Domains](#service-domains)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Development & Testing](#development--testing)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Architecture Overview

LogiSynapse is designed as a set of loosely coupled microservices, each responsible for a distinct domain:

### Shipment Service

- **gRPC API** for shipment creation, updates, and tracking
- **Business Logic**: Validates requests, integrates with Shippo API, persists to Postgres, publishes Kafka events
- **Entrypoint**: `cmd/main.go`
- **Config**: Reads DB and Shippo API keys from environment
- **Persistence**: `store/postgres.go` implements CRUD for shipments
- **Eventing**: Publishes `shipment.created` events via Kafka

### Workflow Orchestrator

- **Temporal Worker**: Executes long-running workflows (e.g., shipment creation, billing)
- **Activities**: Shippo API calls, DB persistence, Kafka publishing
- **Workflow Definitions**: `internal/workflow/create_shipment_workflow.go`
- **Entrypoint**: `cmd/main.go` with dependency injection
- **Retry Logic**: Exponential backoff, activity timeouts, audit trail

### GraphQL Gateway

- **GraphQL API**: Exposes unified API for clients
- **Resolvers**: Map GraphQL requests to gRPC calls
- **Entrypoint**: `cmd/main.go`
- **Schema**: Defined in `graph/schema/schema.graphqls`
- **Client**: gRPC client for shipment-service
- **Models**: Internal models for transport and conversion

### Communications Service

- **Kafka Consumer**: Listens for shipment events
- **RabbitMQ Producer**: Dispatches notification tasks (email, SMS)
- **Worker Pools**: Concurrent processing for notifications
- **Entrypoint**: `cmd/main.go`

### Billing Service

- **Accounts & Payments**: Stripe integration, account management
- **Usage Aggregation**: Tracks billable actions, concurrency-safe
- **Ledger & Invoicing**: Complete financial audit trail, invoice generation, state management
- **Pricing Engine**: Tiered and historical pricing rules
- **API**: Read-only endpoints for usage, invoices, ledger
- **Config**: Secure loading of secrets and keys

---

## Service Domains

### Shipment Service

- Handles shipment lifecycle: create, update, cancel, track
- Integrates with Shippo for label generation and tracking
- Publishes events for downstream consumers
- Ensures idempotency and auditability

### Workflow Orchestrator

- Orchestrates multi-step business processes
- Ensures atomicity and reliability across external APIs, DB, and eventing
- Provides workflow history and compensation logic

### GraphQL Gateway

- Presents a type-safe, client-friendly API
- Handles request validation, enum conversion, and error mapping
- Bridges frontend and backend services

### Communications Service

- Consumes shipment events
- Translates events into notification tasks
- Supports email and SMS delivery via RabbitMQ
- Graceful shutdown and worker management

### Billing Service

- Manages tenant accounts, payment methods, and Stripe customers
- Aggregates usage events for billing
- Maintains a ledger with full transaction metadata
- Generates invoices with legal compliance (quantity, unit price, currency)
- Enforces state transitions and immutability for finalized invoices
- Exposes billing API for usage, invoices, and ledger views

---

## Project Structure

```
LogiSynapse/
├── services/
│   ├── shipment-service/
│   ├── workflow-orchestrator/
│   ├── graphql-gateway/
│   ├── communications-service/
│   └── billing-service/
│       ├── db/migrations/
│       ├── internal/
│       │   ├── accounts/
│       │   ├── billing/
│       │   ├── billingTypes/
│       │   ├── config/
│       │   ├── invoice/
│       │   ├── ledger/
│       │   ├── payment/
│       │   ├── pricing/
│       │   ├── store/
│       │   └── usage/
├── shared/
│   ├── proto/
│   ├── contracts/
│   ├── config/
│   ├── kafka/
│   └── rabbitmq/
├── doc/
├── docker-compose.yml
├── projectsummary.md
├── implementation-plan.md
├── LICENSE
└── README.md
```

---

## Getting Started

### Prerequisites

- **Go 1.24+**
- **Docker & Docker Compose**
- **Shippo API Key** (for shipment service)
- **Stripe API Key** (for billing service)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/Tanmoy095/LogiSynapse.git
   cd LogiSynapse
   ```
2. **Create a `.env` file** in the root (see projectsummary.md for required variables)
3. **Start all services**
   ```bash
   docker-compose up --build
   ```
4. **Access services**
   - GraphQL Playground: http://localhost:8080/
   - Temporal Web UI: http://localhost:8088/
   - RabbitMQ Management: http://localhost:15672/
   - Shipment Service (gRPC): localhost:50051
   - Billing Service: see internal API docs

---

## API Documentation

### GraphQL API (Port 8080)

- **Mutations**: `createShipment`, `updateShipment`, etc.
- **Queries**: `shipments`, `usageSummary`, `invoiceHistory`

### gRPC API (Port 50051)

- **ShipmentService**: `GetShipments`, `CreateShipment`
- **BillingService**: `GetInvoices`, `CreateInvoice`, `FinalizeInvoice`, etc.

### Billing API (Internal)

- **Usage Summary**: Aggregated usage per tenant/month/type
- **Invoice History**: List and details of invoices
- **Ledger View**: Transaction-level audit trail

---

## Development & Testing

- **Run Unit Tests**: `go test ./...`
- **Database Migrations**: Managed via Goose; see `db/migrations/`
- **Code Generation**: Regenerate proto and GraphQL code after schema changes
- **Mocking & Interfaces**: Consumer-defined interfaces for easy mocking in tests
- **Healthchecks**: All services implement readiness checks for orchestration
- **Error Handling**: Structured error mapping and logging throughout

---

## Roadmap

- Add Kafka broker to docker-compose
- Expand test coverage (unit, integration, E2E)
- Implement structured logging and distributed tracing
- Add authentication/authorization to GraphQL Gateway
- Enhance billing API with write endpoints and webhook support
- Add CI/CD pipeline and Makefile for common tasks
- Improve observability and metrics

---

## Contributing

Contributions are welcome! Please follow conventional commit messages and add tests for new features. See projectsummary.md for architectural guidelines.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ using Go, Temporal.io, and modern microservices patterns**
