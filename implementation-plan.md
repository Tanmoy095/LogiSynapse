👨‍🏫 Your Logisynapse 2026 Learning Roadmap
Module 1: Service Resilience & Workflows (Phases 1-4)

Objective: Fix the fragility in your current CreateShipment function. You will learn to manage long-running, distributed transactions and ensure your system is fault-tolerant.

Phase 1: Set Up Temporal

Task: Get the Temporal server running. You don't need to write any code yet. Just add the Temporal server (and its dependencies like Cassandra/Postgres) to your docker-compose.yml file.

Skill Gained: Service-level dependency management.

Phase 2: Create the workflow-orchestrator Service

Task: Create the new services/workflow-orchestrator Go service. Inside, define Temporal Activities. An Activity is just a Go function.

Your first three activities will be the exact business logic you already wrote in shipment.service.go:

Activity_CallShippoAPI(...)

Activity_SaveShipmentToDB(...)

Activity_PublishKafkaEvent(...)

Skill Gained: Defining Temporal Activities; extracting business logic.

Phase 3: Define the CreateShipmentWorkflow

Task: Inside the workflow-orchestrator, define a Temporal Workflow. This workflow function will call your three Activities in sequence, with retries.

Code: err := workflow.ExecuteActivity(ctx, "Activity_CallShippoAPI", ...).Get(ctx, &result)

Skill Gained: Workflow-as-code; managing state, retries, and sequential execution in a distributed system.

Phase 4: Refactor shipment-service

Task: This is the "Aha!" moment. Go back to your services/shipment-service/handler/grpc/shipment.handler.grpc.go.

Delete all the synchronous logic from your CreateShipment handler.

Replace it with a single call to the Temporal client to start your new CreateShipmentWorkflow asynchronously.

Result: Your gRPC handler is now instant. The complex, 3-step, failure-prone logic is now managed by Temporal, which will retry for days if necessary. You have just built a truly resilient service.

Module 2: Full Event-Driven Architecture (Phases 5-7)

Objective: Master the full event-driven pattern by both consuming and publishing events, and learn to select the right tool (Kafka vs. RabbitMQ) for the job.

Phase 5: Implement Kafka Consumption

Task: Fulfill our plan to consolidate status-tracker. In your services/shipment-service (in cmd/main.go), add a Kafka consumer (using segmentio/kafka-go's Reader).

Listen to a new topic (e.g., carrier.status.updates) and have it call a new method s.service.UpdateShipmentStatus(...).

Skill Gained: Kafka consumption; service consolidation; data ownership.

Phase 6: Add RabbitMQ & communications-service

Task: Add RabbitMQ to docker-compose.yml. Create the new services/communications-service.

Why? To learn the difference: Kafka is a durable event log (good for shipment.created). RabbitMQ is a task queue (good for "send this email, I don't care about the event in 3 days").

Skill Gained: Differentiating message brokers; building a dedicated notification service.

Phase 7: Wire RabbitMQ to a Workflow

Task: Update your CreateShipmentWorkflow (Phase 3). Add a fourth step: Activity_QueueWelcomeEmail(...).

This new activity will publish a message to a RabbitMQ queue (e.g., tasks.email.send).

Your communications-service will consume from this queue and "send" the email (i.e., log it).

Skill Gained: Integrating multiple messaging systems; asynchronous task queuing.

Module 3: Enterprise Security & Multi-Tenancy (Phases 8-10)

Objective: Secure your platform. This module is critical for any senior role.

Phase 8: Authentication (AuthN)

Task: Build the services/auth-service. Implement OAuth2/JWT token generation. Add middleware to your services/graphql-gateway that requires a valid JWT.

Skill Gained: Implementing OAuth2/JWT; securing public-facing gateways.

Phase 9: Multi-Tenancy & Postgres RLS

Task: This is a big one.

ALTER TABLE shipments ADD COLUMN tenant_id UUID;

Pass the tenant_id from the JWT (Phase 8) through the gRPC context to the shipment-service.

In your store/postgres.go, run SET app.current_tenant_id = '...' before your query.

Enable Row-Level Security (RLS) on the shipments table in Postgres.

Result: Postgres itself now enforces that no query can ever see or touch data from another tenant.

Skill Gained: Advanced database security; Postgres RLS; secure gRPC metadata handling.

Phase 10: Authorization (AuthZ) with OPA

Task: Build the services/opa-service. In your graphql-gateway middleware, after checking the JWT, make a gRPC call to OPA.

Query: "Does this user (Subject) have permission to create (Action) on this resource shipment (Object)?"

Skill Gained: Decoupled authorization; policy-as-code; Open Policy Agent (OPA).

Module 4: CI/CD & Infrastructure as Code (Phases 11-13)

Objective: Automate your testing and deployment.

Phase 11: Integration Testing with Testcontainers

Task: Your projectsummary.md noted a lack of tests. Let's fix that.

Write a Go integration test for services/shipment-service that uses Testcontainers to programmatically start a real Postgres database. The test will run your migrations, call CreateShipment, and query the DB to verify the result.

Skill Gained: High-fidelity integration testing; Testcontainers.

Phase 12: Build the CI Pipeline

Task: Create your GitHub Actions workflow (.github/workflows/ci.yml).

This pipeline will: Run go test -v ./... (including your new Testcontainers test), run golangci-lint, build all your Docker images, and push them to ECR/Docker Hub.

Skill Gained: CI/CD pipeline automation; GitHub Actions.

Phase 13: Infrastructure as Code (Terraform)

Task: Stop using docker-compose.yml for "prod." Write Terraform scripts to define your production infrastructure (EKS cluster, RDS for Postgres, Kafka, etc.) on AWS.

Skill Gained: Infrastructure as Code (IaC); Terraform; EKS.

Module 5: Observability (The 3 Pillars) (Phases 14-16)

Objective: Make your complex distributed system debuggable.

Phase 14: Pillar 1: Metrics

Task: Instrument all your Go services with OpenTelemetry to export metrics (e.g., gRPC requests, DB call latency). Set up Prometheus & Grafana to scrape and visualize these metrics.

Skill Gained: Metrics; OpenTelemetry; Prometheus; Grafana.

Phase 15: Pillar 2: Tracing

Task: Use OpenTelemetry to add distributed tracing.

Result: You will be able to see a single CreateShipment request flow from the graphql-gateway -> shipment-service -> workflow-orchestrator -> Activity_CallShippoAPI all in one beautiful flame graph in Jaeger.

Skill Gained: Distributed tracing; Jaeger.

Phase 16: Pillar 3: Logging

Task: Configure all Go services to use a structured logger (e.g., slog) to output JSON. Set up Loki to aggregate all service logs and make them searchable in Grafana.

Skill Gained: Structured logging; log aggregation; Loki.

Module 6: Advanced AI & Service Mesh (Phases 17-18)

Objective: Implement the "2026" skills: AI and a service mesh.

Phase 17: Deploy Istio Service Mesh

Task: Install Istio onto your EKS cluster (from Phase 13). Configure it to automatically enforce mTLS (mutual TLS) for all internal gRPC traffic between your services.

Result: All service-to-service communication is now encrypted, and you didn't have to write a single line of Go.

Skill Gained: Service mesh; Istio; mTLS.

Phase 18: Build the RAG/AI Pipeline

Task: This is a multi-part phase.

Build services/doc-intelligence-service.

Set up the pgvector extension in your Postgres (RDS).

Have doc-intelligence consume shipment.created events from Kafka, create vector embeddings from the shipment data, and store them in pgvector.

Build services/ai-analysis-service with a gRPC endpoint AnswerQuestion(...).

This endpoint will take a question, query pgvector, and pass the context to the Google Gemini API to get an answer.

Skill Gained: Retrieval-Augmented Generation (RAG); pgvector; LLM SDK integration.

Module 7: The Platform Engineering Frontier (Phases 19-20)

Objective: Master the most advanced skills that differentiate a senior engineer from a platform architect.

Phase 19: Chaos Engineering with Istio

Task: Use Istio's "Fault Injection" feature. Purposefully inject a 500ms latency spike or a 10% error rate on calls to the shipment-service.

Verify: Watch your Temporal dashboard (from Phase 1) and your Grafana dashboard (from Phase 14). Prove that your workflow retries and your alerts fire correctly.

Skill Gained: Chaos engineering; validating system resilience.

Phase 20: Extensibility with WebAssembly (WASM)

Task: Build the final service: services/anomaly-plugin-service.

Use Wazero (a pure-Go WASM runtime) to create a service that can securely run customer-supplied logic (as a WASM file) in a sandbox.

Skill Gained: WebAssembly (WASM) in Go; sandboxed plugin systems; Wazero.

This 20-phase plan
