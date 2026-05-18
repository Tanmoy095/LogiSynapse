# LogiSynapse

> Modern, event-driven shipment and billing platform with distributed workflow orchestration.

LogiSynapse is an AI-native distributed logistics intelligence platform that combines production-grade backend engineering with practical AI systems engineering.

LogiSynapse is a logistics operating system for merchants, operators, support teams, finance teams, and engineering teams. It manages the full shipment lifecycle: merchant order -> shipment creation -> carrier selection -> label generation -> tracking updates -> delay and exception handling -> customer notifications -> usage billing -> operational analytics -> AI-assisted decisions.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Service Domains](#service-domains)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Development and Testing](#development-and-testing)
- [Roadmap Highlights](#roadmap-highlights)
- [Contributing](#contributing)
- [License](#license)

## Architecture Overview

LogiSynapse is designed around clear, loosely coupled planes that separate product concerns, data ownership, orchestration, and intelligence.

### Key Features

- Event-driven architecture with transactional outbox and Kafka for replayable domain events
- Durable workflow orchestration using Temporal for long-running, retryable processes
- AI-native capabilities including retrieval, RAG, typed tools, and audited AI workflows
- Microservices design with clear bounded contexts for order, tracking, billing, notifications, and AI
- Observability and auditability with OpenTelemetry, metrics, traces, and a full audit trail
- Production patterns such as idempotency, retries, compensation, and strong testing boundaries

### Architecture Planes

- Product plane: API gateway, order, tracking, support, dispatch
- Data plane: PostgreSQL, Redis, Kafka, vector DB for durable facts and read models
- Workflow plane: Temporal workers and orchestrators for durable business processes
- Intelligence plane: AI gateway, model gateway, retrieval, tool-service, and eval pipelines

### Core Flow Examples

- Order write: API -> order-service -> Postgres + outbox -> outbox relay -> Kafka -> downstream consumers
- Tracking read: API -> tracking-service -> Redis cache -> Postgres fallback
- AI assistant: ai-gateway -> retrieval -> model -> typed tools -> validated, cited response

## Service Domains

| Service | Responsibility | Primary Integrations |
|---|---|---|
| api-gateway | Public API surface, auth, rate limits | GraphQL / REST |
| order-service | Accept orders, outbox, idempotency | Postgres, Kafka |
| tracking-service | Shipment timeline and read models | Redis, Postgres |
| workflow-service | Temporal workflows and retries | Temporal, Shippo |
| billing-service | Usage aggregation, ledger, invoices | Stripe, Postgres |
| notification-service | Email, SMS, webhooks | RabbitMQ, SQS |
| ai-gateway | Tenant AI requests, quotas, streaming | model-gateway, retrieval |
| retrieval-service | Embeddings and hybrid search | pgvector / Qdrant |

## Project Structure

```text
LogiSynapse/
├── services/
│   ├── shipment-service/
│   ├── workflow-orchestrator/
│   ├── graphql-gateway/
│   ├── communications-service/
│   └── billing-service/
├── shared/
├── docs/
├── graphify-out/
├── docker-compose.yml
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.24+
- Docker and Docker Compose
- Shippo API key for shipment workflows
- Stripe API key for billing workflows

### Quick Start

1. Clone the repository.

	```bash
	git clone https://github.com/Tanmoy095/LogiSynapse.git
	cd LogiSynapse
	```

2. Create a `.env` file in the project root with the required database and integration settings.

3. Start the local stack.

	```bash
	docker compose up --build
	```


### GraphQL API

- Mutations: `createShipment`, `updateShipment`, and related workflow actions
- Queries: `shipments`, `usageSummary`, `invoiceHistory`

### gRPC API

- ShipmentService: `CreateShipment`, `GetShipments`
- BillingService: `GetInvoices`, `CreateInvoice`, `FinalizeInvoice`

### Billing API

- Usage summary by tenant, period, and type
- Invoice history and invoice details
- Ledger views for transaction-level auditability

## Development and Testing

- Run tests:

  ```bash
  go test ./...
  ```

- Database migrations are managed with Goose under service migration folders.
- Regenerate proto and GraphQL artifacts after schema changes.
- Keep health checks and observability hooks in place for each service.
- Use OpenTelemetry for traces and metrics across the stack.

## Roadmap Highlights

- Add Kafka to local compose for end-to-end event testing
- Expand unit, integration, and E2E coverage
- Strengthen logging, tracing, and AI evaluation pipelines
# LogiSynapse

> Modern, event-driven shipment and billing platform with distributed workflow orchestration.

LogiSynapse is an AI-native distributed logistics intelligence platform that combines production-grade backend engineering with practical AI systems engineering.

LogiSynapse is a logistics operating system for merchants, operators, support teams, finance teams, and engineering teams. It manages the full shipment lifecycle: merchant order -> shipment creation -> carrier selection -> label generation -> tracking updates -> delay and exception handling -> customer notifications -> usage billing -> operational analytics -> AI-assisted decisions.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Service Domains](#service-domains)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Development and Testing](#development-and-testing)
- [Roadmap Highlights](#roadmap-highlights)
- [Contributing](#contributing)
- [License](#license)

## Architecture Overview

LogiSynapse is designed around clear, loosely coupled planes that separate product concerns, data ownership, orchestration, and intelligence.

### Key Features

- Event-driven architecture with transactional outbox and Kafka for replayable domain events
- Durable workflow orchestration using Temporal for long-running, retryable processes
- AI-native capabilities including retrieval, RAG, typed tools, and audited AI workflows
- Microservices design with clear bounded contexts for order, tracking, billing, notifications, and AI
- Observability and auditability with OpenTelemetry, metrics, traces, and a full audit trail
- Production patterns such as idempotency, retries, compensation, and strong testing boundaries

### Architecture Planes

- Product plane: API gateway, order, tracking, support, dispatch
- Data plane: PostgreSQL, Redis, Kafka, vector DB for durable facts and read models
- Workflow plane: Temporal workers and orchestrators for durable business processes
- Intelligence plane: AI gateway, model gateway, retrieval, tool-service, and eval pipelines

### Core Flow Examples

- Order write: API -> order-service -> Postgres + outbox -> outbox relay -> Kafka -> downstream consumers
- Tracking read: API -> tracking-service -> Redis cache -> Postgres fallback
- AI assistant: ai-gateway -> retrieval -> model -> typed tools -> validated, cited response

## Service Domains

| Service | Responsibility | Primary Integrations |
|---|---|---|
| api-gateway | Public API surface, auth, rate limits | GraphQL / REST |
| order-service | Accept orders, outbox, idempotency | Postgres, Kafka |
| tracking-service | Shipment timeline and read models | Redis, Postgres |
| workflow-service | Temporal workflows and retries | Temporal, Shippo |
| billing-service | Usage aggregation, ledger, invoices | Stripe, Postgres |
| notification-service | Email, SMS, webhooks | RabbitMQ, SQS |
| ai-gateway | Tenant AI requests, quotas, streaming | model-gateway, retrieval |
| retrieval-service | Embeddings and hybrid search | pgvector / Qdrant |

## Project Structure

```text
LogiSynapse/
├── services/
│   ├── shipment-service/
│   ├── workflow-orchestrator/
│   ├── graphql-gateway/
│   ├── communications-service/
│   └── billing-service/
├── shared/
├── docs/
├── graphify-out/
├── docker-compose.yml
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.24+
- Docker and Docker Compose
- Shippo API key for shipment workflows
- Stripe API key for billing workflows

### Quick Start

1. Clone the repository.

	```bash
	git clone https://github.com/Tanmoy095/LogiSynapse.git
	cd LogiSynapse
	```

2. Create a `.env` file in the project root with the required database and integration settings.

3. Start the local stack.

	```bash
	docker compose up --build
	```

4. Open the local tools.

	- GraphQL Playground: http://localhost:8080/
	- Temporal Web UI: http://localhost:8088/
	- RabbitMQ Management: http://localhost:15672/

## API Documentation

### GraphQL API

- Mutations: `createShipment`, `updateShipment`, and related workflow actions
- Queries: `shipments`, `usageSummary`, `invoiceHistory`

### gRPC API

- ShipmentService: `CreateShipment`, `GetShipments`
- BillingService: `GetInvoices`, `CreateInvoice`, `FinalizeInvoice`

### Billing API

- Usage summary by tenant, period, and type
- Invoice history and invoice details
- Ledger views for transaction-level auditability

## Development and Testing

- Run tests:

  ```bash
  go test ./...
  ```

- Database migrations are managed with Goose under service migration folders.
- Regenerate proto and GraphQL artifacts after schema changes.
- Keep health checks and observability hooks in place for each service.
- Use OpenTelemetry for traces and metrics across the stack.

## Roadmap Highlights

- Add Kafka to local compose for end-to-end event testing
- Expand unit, integration, and E2E coverage
- Strengthen logging, tracing, and AI evaluation pipelines
- Harden the gateway with authentication and authorization
- Add CI/CD, Makefile tasks, and deployment manifests

## Contributing

- Follow conventional commits.
- Keep changes small and focused on one responsibility.
- Add tests and documentation for new behavior.
- Document tradeoffs and failure modes for architecture changes.

## License

GPLv3 - see the [LICENSE](LICENSE) file for details.
