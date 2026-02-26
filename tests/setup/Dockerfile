FROM golang:1.24-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/test-service ./tests/setup/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/test-service /usr/local/bin/test-service

EXPOSE 5679

CMD ["/usr/local/bin/test-service"]
