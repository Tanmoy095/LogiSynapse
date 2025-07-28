# syntax=docker/dockerfile:1

# ------------ Builder Stage ------------ #
FROM golang:1.24 AS builder

# Install PostgreSQL client for pg_isready
RUN apt-get update && apt-get install -y postgresql-client

# Set working directory
WORKDIR /app

# Copy Go modules for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Install goose for DB migration (will be copied to final image)
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Build the application binary
RUN go build -o shipment-service ./cmd

# ------------ Final Stage (Slim Runtime) ------------ #
FROM debian:bookworm-slim

# Install minimal PostgreSQL client
RUN apt-get update && apt-get install -y postgresql-client ca-certificates curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/shipment-service .
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy scripts and make them executable
COPY wait-for-postgres.sh entrypoint.sh ./
RUN chmod +x wait-for-postgres.sh entrypoint.sh

# Migrations folder
COPY db/migrations /migrations


# Expose port (optional if you're serving HTTP)
EXPOSE 8080

# Entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
