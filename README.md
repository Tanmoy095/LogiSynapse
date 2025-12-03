# LogiSynapse — Microservices-Based Shipment Management Platform

LogiSynapse is a modern, scalable shipment management platform built with **Go**, **gRPC**, **GraphQL**, **Temporal.io**, and **PostgreSQL**. It orchestrates complex logistics workflows including shipment creation, tracking, label generation, and carrier rate comparison through integration with **Shippo's API**. The architecture follows microservices patterns with event-driven communication via Kafka, inspired by enterprise systems like Amazon and FedEx.

## 🏗️ Architecture Overview

LogiSynapse consists of three core services working together:

### 1. **Shipment Service** (gRPC API)

The core domain service handling all shipment operations via gRPC:

- **CRUD Operations**: Create, Read, Update, Delete shipments
- **External Integration**: Shippo API for label generation and tracking
- **Data Persistence**: PostgreSQL with automated migrations
- **Event Publishing**: Kafka events for shipment lifecycle changes
- **Status Management**: Proto-based enum system for shipment states

### 2. **Workflow Orchestrator** (Temporal Worker)

Orchestrates multi-step shipment workflows using **Temporal.io**:

- **Reliable Execution**: Durable workflows with automatic retries and exponential backoff
- **Activity Decomposition**:
  - `ACTIVITY_CallShippoAPI` — External API calls to Shippo
  - `ACTIVITY_SaveShipmentToDB` — Database persistence
  - `ACTIVITY_PublishKafkaEvent` — Event publication
- **Saga Pattern**: Handles distributed transactions and compensations
- **Fault Tolerance**: Survives service restarts and network failures

### 3. **GraphQL Gateway** (Client-Facing API)

User-facing API layer exposing shipment operations via GraphQL:

- **Schema-First Design**: Type-safe GraphQL schema
- **gRPC Client**: Communicates with Shipment Service
- **Query & Mutations**: Read and write operations for shipments
- **API Aggregation**: Single entry point for frontend applications

### Supporting Infrastructure

- **Shared Module**: Common contracts, proto definitions, Kafka producers, and config
- **Temporal Server**: Workflow engine with Web UI (port 8088)
- **PostgreSQL**: Two databases (app data + Temporal state)
- **Docker Compose**: Full-stack orchestration

## 📊 Current Implementation Status

### ✅ Completed Features

#### Core Shipment Operations

- ✅ Create shipments with dynamic package dimensions (length, width, height, weight)
- ✅ Update shipment details (destination, ETA) for `PRE_TRANSIT` status
- ✅ Cancel shipments with Shippo API integration for label voiding
- ✅ Get single shipment by ID
- ✅ List all shipments with pagination support
- ✅ Carrier rate comparison (FedEx, UPS, DHL)
- ✅ Shipping label generation via Shippo

#### Temporal Workflow Integration

- ✅ `CreateShipmentWorkflow` with 3-step orchestration
- ✅ Retry policies with exponential backoff (up to 10 minutes)
- ✅ Activity timeout configuration (45 seconds per step)
- ✅ Dependency injection for Store and Kafka Producer
- ✅ Worker registration with proper activity/workflow setup

#### Data & Schema Management

- ✅ Centralized proto definitions in `shared/proto/`
- ✅ Unified `ShipmentStatus` enum (PENDING, PRE_TRANSIT, IN_TRANSIT, DELIVERED, CANCELLED)
- ✅ Shared contracts in `shared/contracts/` for cross-service models
- ✅ Database migrations with `pgcrypto` extension
- ✅ Shipments table with tracking, carrier, dimensions fields

#### Infrastructure & DevOps

- ✅ Multi-service Docker Compose setup
- ✅ Temporal Server with dedicated PostgreSQL database
- ✅ Temporal Web UI (accessible at `http://localhost:8088`)
- ✅ Health checks for all services
- ✅ Shared network (`loginet`) for inter-service communication
- ✅ Environment-based configuration via `.env`

#### Code Quality & Organization

- ✅ Monorepo structure with Go workspaces
- ✅ Local `replace` directives for inter-module dependencies
- ✅ Centralized config loading (`shared/config/config.go`)
- ✅ Kafka producer abstraction (`shared/kafka/producer.go`)
- ✅ Proto-based gRPC contracts with generated code

### 🚧 In Progress / Planned Features

#### Webhook Integration

- 🔄 Shippo webhook receiver for real-time tracking updates
- 🔄 Automatic status synchronization on carrier events

#### Enhanced Workflows

- 🔄 UpdateShipment workflow via Temporal
- 🔄 CancelShipment workflow with compensation logic
- 🔄 Rate comparison workflow with caching

#### Observability

- 🔄 Structured logging (zerolog/zap)
- 🔄 Distributed tracing (OpenTelemetry)
- 🔄 Metrics collection (Prometheus)
- 🔄 Temporal workflow monitoring dashboards

#### Testing

- 🔄 Unit tests for activities and workflows
- 🔄 Integration tests with Temporal test server
- 🔄 Mock Shippo API for local testing
- 🔄 End-to-end GraphQL query tests

## 🗂️ Project Structure

```
LogiSynapse/
├── services/
│   ├── shipment-service/          # Core gRPC service
│   │   ├── cmd/main.go            # Service entrypoint
│   │   ├── handler/grpc/          # gRPC handlers
│   │   ├── service/               # Business logic
│   │   ├── store/                 # Database layer (Postgres)
│   │   ├── db/migrations/         # SQL migrations
│   │   └── Dockerfile
│   │
│   ├── workflow-orchestrator/     # Temporal worker
│   │   ├── cmd/main.go            # Worker entrypoint
│   │   ├── internal/
│   │   │   ├── activities/        # Temporal activities
│   │   │   └── workflow/          # Workflow definitions
│   │   └── go.mod
│   │
│   └── graphql-gateway/           # GraphQL API
│       ├── cmd/main.go            # Gateway entrypoint
│       ├── graph/                 # GraphQL resolvers & schema
│       ├── client/                # gRPC client
│       └── Dockerfile
│
├── shared/                        # Shared modules
│   ├── proto/                     # Protobuf definitions
│   │   ├── shipment.proto         # Service contract
│   │   ├── shipment.pb.go         # Generated Go code
│   │   └── shipment_grpc.pb.go    # gRPC stubs
│   ├── contracts/                 # Domain models
│   │   └── shipment.model.go      # Shared Shipment struct
│   ├── config/                    # Configuration utilities
│   │   └── config.go              # CommonConfig loader
│   ├── kafka/                     # Kafka abstractions
│   │   └── producer.go            # Publisher interface
│   └── go.mod
│
├── docker-compose.yml             # Full-stack orchestration
├── .env                           # Environment variables
└── README.md                      # This file
```

## 🔄 Data Flow: Creating a Shipment

### Option 1: Via Temporal Workflow (Recommended)

```
[GraphQL Client]
      │
      ├─→ [GraphQL Gateway]
      │         │
      │         ├─→ Starts Temporal Workflow
      │         │         │
      │         │         ├─→ [Workflow Orchestrator]
      │         │         │         │
      │         │         │         ├─→ ACTIVITY: Call Shippo API
      │         │         │         │         └─→ Returns tracking number
      │         │         │         │
      │         │         │         ├─→ ACTIVITY: Save to Postgres
      │         │         │         │         └─→ Stores shipment
      │         │         │         │
      │         │         │         └─→ ACTIVITY: Publish Kafka Event
      │         │         │                   └─→ shipment.created
      │         │         │
      │         │         └─→ Returns result
      │         │
      │         └─→ Returns to client
```

### Option 2: Direct gRPC (Legacy Path)

```
[GraphQL Client]
      │
      └─→ [GraphQL Gateway]
               │
               └─→ [Shipment Service] (gRPC)
                        │
                        ├─→ Validates request
                        ├─→ Calls Shippo API (HTTP)
                        ├─→ Saves to Postgres
                        ├─→ Publishes Kafka event
                        └─→ Returns proto response
```

## 🛠️ Technology Stack

| Layer                | Technology                 | Purpose                     |
| -------------------- | -------------------------- | --------------------------- |
| **API Gateway**      | GraphQL (gqlgen)           | Client-facing API           |
| **Service Layer**    | gRPC (Go)                  | Inter-service communication |
| **Workflow Engine**  | Temporal.io                | Orchestration & reliability |
| **Database**         | PostgreSQL 15              | Data persistence            |
| **Messaging**        | Kafka (segmentio/kafka-go) | Event streaming             |
| **External API**     | Shippo REST API            | Shipping & tracking         |
| **Containerization** | Docker + Docker Compose    | Local development           |
| **Protocol**         | Protocol Buffers           | Service contracts           |

## 🚀 Getting Started

### Prerequisites

- Go 1.24.4+
- Docker & Docker Compose
- Shippo API Key (sign up at [goshippo.com](https://goshippo.com))

### Environment Setup

1. **Clone the repository** (with permission):

```bash
git clone https://github.com/Tanmoy095/LogiSynapse.git
cd LogiSynapse
```

2. **Create `.env` file** in the root:

```env
# Database
DB_USER=postgres
DB_PASSWORD=your_password
DB_HOST=postgres
DB_PORT=5432
DB_NAME=logisyncdb

# Shippo API
SHIPPO_API_KEY=shippo_test_your_key_here

# Kafka
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC=shipment-events

# Temporal
TEMPORAL_HOST_PORT=temporal:7233
TEMPORAL_VERSION=1.24.2
TEMPORAL_UI_VERSION=2.26.2
```

3. **Start all services**:

```bash
docker-compose up --build
```

4. **Access the services**:

- GraphQL Playground: `http://localhost:8080`
- Temporal Web UI: `http://localhost:8088`
- Shipment Service (gRPC): `localhost:50051`
- PostgreSQL: `localhost:5432`
- Temporal DB: `localhost:5433`

### Running Services Locally (Development)

#### Shipment Service

```bash
cd services/shipment-service
go mod tidy
export SHIPPO_API_KEY="your_key"
export DB_HOST="localhost"
# ... other env vars
go run cmd/main.go
```

#### Workflow Orchestrator

```bash
cd services/workflow-orchestrator
go mod tidy
export TEMPORAL_HOST_PORT="localhost:7233"
# ... other env vars
go run cmd/main.go
```

#### GraphQL Gateway

```bash
cd services/graphql-gateway
go mod tidy
export SHIPMENT_SERVICE_ADDR="localhost:50051"
go run cmd/main.go
```

## 📝 Key Implementation Details

### Temporal Workflow Architecture

#### Workflow Definition (`create_shipment_workflow.go`)

The `CreateShipmentWorkflow` orchestrates three activities in sequence:

```go
func CreateShimentWorkflow(ctx workflow.Context, shipment contracts.Shipment) (contracts.Shipment, error) {
    // Step 1: Call Shippo API
    var shippoResult contracts.Shipment
    workflow.ExecuteActivity(ctx, "ACTIVITY_CallShippoAPI", shipment).Get(ctx, &shippoResult)

    // Step 2: Save to Database
    var storedShipment contracts.Shipment
    workflow.ExecuteActivity(ctx, "ACTIVITY_SaveShipmentToDB", shippoResult).Get(ctx, &storedShipment)

    // Step 3: Publish Kafka Event
    workflow.ExecuteActivity(ctx, "ACTIVITY_PublishKafkaEvent", storedShipment).Get(ctx, nil)

    return storedShipment, nil
}
```

**Retry Configuration**:

- Initial interval: 1 second
- Backoff coefficient: 2.0 (exponential)
- Maximum interval: 1 minute
- Maximum attempts: 100
- Activity timeout: 45 seconds

#### Activity Implementation (`shipment_activities.go`)

**`ACTIVITY_CallShippoAPI`**:

- Validates package dimensions
- Constructs Shippo API request with dynamic parcel data
- Sends HTTP POST to `https://api.goshippo.com/shipments`
- Parses response to extract tracking number, status, label URL
- Maps Shippo status strings to proto enums

**`ACTIVITY_SaveShipmentToDB`**:

- Wraps the store's `CreateShipment` method
- Uses the injected Store interface for testability

**`ACTIVITY_PublishKafkaEvent`**:

- Publishes `shipment.created` event
- Uses injected Kafka Producer interface

#### Worker Setup (`cmd/main.go`)

Dependencies are injected at worker startup:

```go
activityHost := &activities.ShipmentActivities{
    Store:     shipmentStore,               // From shared/config
    Producer:  producer,                    // Kafka producer
    ShippoKey: os.Getenv("SHIPPO_API_KEY"),
    Client:    &http.Client{Timeout: 10 * time.Second},
}

w := worker.New(c, "SHIPMENT_TASK_QUEUE", worker.Options{})
w.RegisterWorkflow(workflow.CreateShimentWorkflow)
w.RegisterActivity(activityHost.ACTIVITY_CallShippoAPI)
// ... register other activities
```

### Shared Module Refactoring

#### Contracts (`shared/contracts/shipment.model.go`)

Single source of truth for domain models:

```go
type Shipment struct {
    ID             string
    Origin         string
    Destination    string
    Eta            string
    Status         proto.ShipmentStatus  // Proto enum!
    Carrier        Carrier
    TrackingNumber string
    Length         float64
    Width          float64
    Height         float64
    Weight         float64
    Unit           string
}
```

**Benefits**:

- No duplication across services
- Proto enum integration eliminates string mismatches
- All services use identical model

#### Config (`shared/config/config.go`)

Centralized infrastructure configuration:

```go
type CommonConfig struct {
    DB_USER      string
    DB_PASSWORD  string
    DB_NAME      string
    DB_HOST      string
    DB_PORT      string
    KAFKA_TOPIC  string
    KAFKA_BROKER string
}

func (c *CommonConfig) GetDBURL() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", ...)
}
```

Both `shipment-service` and `workflow-orchestrator` use this for DB connection.

#### Kafka Abstraction (`shared/kafka/producer.go`)

Interface-based design for testability:

```go
type Publisher interface {
    Publish(ctx context.Context, key string, value interface{}) error
    Close() error
}
```

Activities use the interface, making them unit-testable with mocks.

### Proto Enum System

#### Definition (`shared/proto/shipment.proto`)

```protobuf
enum ShipmentStatus {
  PENDING = 0;
  PRE_TRANSIT = 1;
  IN_TRANSIT = 2;
  DELIVERED = 3;
  CANCELLED = 4;
}
```

#### Status Mapping

Activities map Shippo strings to enums:

```go
func mapShippoStatusToProto(s string) proto.ShipmentStatus {
    switch s {
    case "PRE_TRANSIT":
        return proto.ShipmentStatus_PRE_TRANSIT
    case "IN_TRANSIT":
        return proto.ShipmentStatus_IN_TRANSIT
    // ...
    default:
        return proto.ShipmentStatus_PENDING
    }
}
```

## 🧪 Testing & Development

### Build Verification

```bash
# Test all modules compile
cd /path/to/LogiSynapse
go build ./...

# Specific service
cd services/shipment-service
go build ./cmd
```

### Module Dependencies

Each service uses `replace` directives for local development:

**`services/workflow-orchestrator/go.mod`**:

```go
replace github.com/Tanmoy095/LogiSynapse/shared => ../../shared
replace github.com/Tanmoy095/LogiSynapse/services/shipment-service => ../shipment-service
```

After changes, run:

```bash
cd services/workflow-orchestrator
go mod tidy
```

### Docker Compose Services

**Services running**:

1. `postgres` — Application database (port 5432)
2. `temporal-db` — Temporal's database (port 5433)
3. `temporal` — Temporal server (gRPC on 7233)
4. `temporal-ui` — Web UI (port 8088)
5. `shipment-service` — gRPC API (port 50051)
6. `graphql-gateway` — GraphQL API (port 8080)
7. `workflow-orchestrator` — Temporal worker (no exposed port)

**Healthchecks**:

- Postgres services wait for `pg_isready`
- Shipment service waits for postgres to be healthy
- Temporal waits for temporal-db to be healthy

## 🐛 Troubleshooting

### Common Issues

**1. Go module errors**: `module not found`

```bash
# Solution: Ensure replace directives are correct
cd services/workflow-orchestrator
go mod edit -replace github.com/Tanmoy095/LogiSynapse/shared=../../shared
go mod tidy
```

**2. Temporal connection refused**

```bash
# Solution: Check Temporal server is running
docker-compose ps temporal
docker-compose logs temporal

# Verify TEMPORAL_HOST_PORT matches your setup
# Docker: temporal:7233
# Local: localhost:7233
```

**3. Shippo API errors (401 Unauthorized)**

```bash
# Solution: Verify API key format
echo $SHIPPO_API_KEY
# Should start with: shippo_test_... or shippo_live_...
```

**4. Proto enum vs string mismatches**

```go
// ❌ Wrong
if shipment.Status == "PENDING" { ... }

// ✅ Correct
if shipment.Status == proto.ShipmentStatus_PENDING { ... }

// Converting to string
statusStr := shipment.Status.String()  // "PENDING"
```

**5. Worker not picking up workflows**

```bash
# Check worker logs
docker-compose logs workflow-orchestrator

# Verify task queue name matches
# Worker: "SHIPMENT_TASK_QUEUE"
# Client: Must use same queue name when starting workflow
```

## 📋 Recent Development History

Based on recent git commits:

1. **e9b1157** (Latest) — Implement workflow-orchestrator and refactor shared modules

   - Created full Temporal worker service
   - Moved models to `shared/contracts/`
   - Centralized config in `shared/config/`

2. **b622e1c** — Implement CreateShipment workflow and worker setup

   - Defined 3-step workflow
   - Implemented activities with Shippo integration
   - Added retry policies and timeouts

3. **abe8413** — Centralize shared models into 'contracts' and add Temporal to compose

   - Added Temporal services to docker-compose
   - Created `shared/contracts/` package
   - Eliminated model duplication

4. **0f29e94** — Centralize shipment proto & add testable Kafka producer

   - Moved proto to `shared/proto/`
   - Created `Publisher` interface for Kafka
   - Made activities testable

5. **6bfa063** — Centralize shipment proto, unify ShipmentStatus enum
   - Single proto definition for all services
   - Enum-based status system
   - Eliminated string-based status bugs

### Key Architectural Decisions

**Why Temporal?**

- Ensures shipment creation is atomic across 3 systems (Shippo, DB, Kafka)
- Automatic retries prevent data loss from transient failures
- Workflow history provides audit trail
- Simplifies error handling and compensation logic

**Why Shared Contracts?**

- Single source of truth eliminates sync issues
- Proto enums prevent string typos
- All services compile against same types
- Refactoring becomes safer

**Why Interface-Based Dependencies?**

- Activities can be unit tested with mocks
- Swap implementations without changing workflow code
- Enables local development without Kafka/Postgres

## 📚 Further Documentation

### Key Files to Study

**Understanding Temporal Implementation**:

1. `services/workflow-orchestrator/cmd/main.go` — Worker setup & DI
2. `services/workflow-orchestrator/internal/workflow/create_shipment_workflow.go` — Workflow logic
3. `services/workflow-orchestrator/internal/activities/shipment_activities.go` — Activity implementations

**Understanding Shared Modules**:

1. `shared/contracts/shipment.model.go` — Domain model
2. `shared/config/config.go` — Infrastructure config
3. `shared/proto/shipment.proto` — Service contract
4. `shared/kafka/producer.go` — Event publishing

**Understanding Service Layer**:

1. `services/shipment-service/handler/grpc/shipment.handler.grpc.go` — gRPC handlers
2. `services/shipment-service/service/shipment.service.go` — Business logic
3. `services/shipment-service/store/postgres.go` — Data access

### Next Steps for Development

1. **Add Unit Tests**:

   - Mock Store and Producer interfaces
   - Test activities in isolation
   - Test workflow logic with Temporal test framework

2. **Implement UpdateShipment Workflow**:

   - Similar 3-step pattern
   - Add compensation logic for failures

3. **Add Observability**:

   - Structured logging with context
   - Distributed tracing across services
   - Metrics for workflow execution times

4. **Enhance Error Handling**:
   - Custom error types for different failure modes
   - Better error messages in GraphQL responses
   - Detailed workflow failure reasons in Temporal UI

## 📄 License

LogiSynapse is licensed under the [Creative Commons Attribution 4.0 International License (CC BY 4.0)](https://creativecommons.org/licenses/by/4.0/).

**You may**:

- Share and adapt the material
- Use for commercial purposes

**You must**:

- Give appropriate credit to Tanmoy095
- Include a link to the license
- Indicate if changes were made

See the [LICENSE](LICENSE) file for details.

## ⚠️ Usage Restrictions

This repository is primarily for **viewing and demonstration purposes**.

**Unauthorized activities prohibited without explicit permission**:

- Cloning for production use
- Redistribution without attribution
- Commercial deployment
- Removal of attribution

**Contributions**: Pull requests and issues are not accepted without prior approval.

## 📧 Contact

For inquiries, permission requests, or collaboration:

- GitHub: [@Tanmoy095](https://github.com/Tanmoy095)
- Repository: [LogiSynapse](https://github.com/Tanmoy095/LogiSynapse)

---

**Built with ❤️ using Go, Temporal, gRPC, and GraphQL**
