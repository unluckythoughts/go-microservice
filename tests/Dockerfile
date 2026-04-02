# Build stage
FROM golang:1.26-alpine

WORKDIR /app

RUN apk add --no-cache make

# Copy root module files so the local replace directive resolves
COPY go.mod go.sum ./
COPY *.go ./
COPY tools/ ./tools/
COPY utils/ ./utils/
COPY examples/ ./examples/

# Copy tests and Makefile preserving directory structure
COPY tests/ ./tests/
COPY Makefile .
RUN go mod download

ENTRYPOINT ["make"]
