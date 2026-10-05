# syntax=docker/dockerfile:1
FROM golang:1.25 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /dc-payment-gateway ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /dc-payment-gateway .
EXPOSE 8080
CMD ["./dc-payment-gateway"]