# Use Go 1.24 to match go.mod requirement
FROM golang:1.24 AS builder

# Install PostgreSQL client tools for pg_isready
RUN apt-get update && apt-get install -y postgresql-client

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Build the binary
RUN go build -o shipment-service ./cmd

# Copy and make wait-for-postgres.sh executable
COPY wait-for-postgres.sh .
RUN chmod +x wait-for-postgres.sh

# Command (overridden in docker-compose.yaml)
CMD ["./shipment-service"]