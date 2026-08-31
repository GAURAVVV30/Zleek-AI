FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build all three binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api-server ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/worker-server ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/migrate-runner ./cmd/migrate

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/bin/api-server .
COPY --from=builder /app/bin/worker-server .
COPY --from=builder /app/bin/migrate-runner .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./api-server"]
