# Build stage
FROM golang:1.23.2 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o metrics-server cmd/main.go

# Deploy stage
FROM alpine:latest

RUN apk add --no-cache libc6-compat

WORKDIR /app

COPY --from=builder /app/metrics-server .

RUN chmod +x /app/metrics-server

CMD ["./metrics-server"]