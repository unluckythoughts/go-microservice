# Build stage
FROM golang:1.26-alpine AS builder

# Mirror repo structure so replace ../../ in examples/microservice/go.mod resolves to /app
WORKDIR /app

# Install build dependencies for CGO (sqlite requires it)
RUN apk add --no-cache gcc musl-dev

# Copy root module (the local dependency)
COPY go.mod go.sum ./
COPY *.go ./
COPY tools/ ./tools/
COPY utils/ ./utils/

# Copy microservice module and download dependencies
# WORKDIR matches examples/microservice so replace ../../ -> /app
WORKDIR /app/examples/microservice
COPY examples/microservice/go.mod examples/microservice/go.sum ./
RUN go mod download

# Copy source and build
COPY examples/microservice/migrations ./migrations/
COPY examples/microservice/main.go .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# Runtime stage
FROM alpine:3.21

WORKDIR /app

# ca-certificates for HTTPS outbound calls; tzdata for timezone support; curl for healthcheck
RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

ENTRYPOINT ["/app/server"]
