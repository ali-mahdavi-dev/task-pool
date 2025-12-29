# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o task-pool ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/task-pool .

EXPOSE 8080

CMD ["./task-pool", "http"]

