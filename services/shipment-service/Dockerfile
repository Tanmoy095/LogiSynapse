#Use go 1.24 to match go.mod

FROM golang:1.24 AS builder

#install postgresSQL client-tool  for pg is_ready
RUN apt-get update && apt-get install -y postgresql-client

#set working directory
WORKDIR /app

#Copy go.mod go.sum for dependency catching..

COPY go.mod go.sum ./
RUN go mod download      
#download dependencies to save times in future buildsw
#copy source code
COPY . .
#build the binary
RUN go build -o shipment-service ./cmd

# copy and make wait for postgres.sh
COPY wait-for-postgres.sh .
RUN chmod +x wait-for-postgres.sh

#Command (ovverridden in docker-compose.yml)
CMD ["./shipment-service"]





