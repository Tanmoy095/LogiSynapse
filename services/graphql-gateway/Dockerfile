FROM golang:1.24 AS builder

# Copy the entire repository into the builder so local module replace paths (../../shared)
# resolve during `go mod download` when building inside the container.
WORKDIR /src
COPY . .

# Switch to the graphql-gateway module and download dependencies
WORKDIR /src/services/graphql-gateway
RUN go mod download

# Build the gateway binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/graphql-gateway ./cmd

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/graphql-gateway ./graphql-gateway

EXPOSE 8080
CMD ["./graphql-gateway"]