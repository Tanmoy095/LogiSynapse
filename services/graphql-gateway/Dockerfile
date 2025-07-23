# Use Go 1.24 to match go.mod requirement
FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o graphql-gateway ./cmd

CMD ["./graphql-gateway"]