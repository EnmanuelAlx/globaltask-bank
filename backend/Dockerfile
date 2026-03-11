# ==========================================
# Base Stage (Shared)
# ==========================================
FROM golang:1.25-alpine AS base

# Install dependencies
RUN apk add --no-cache git gcc musl-dev
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# ==========================================
# Development Stage (Hot Reloading with Air)
# ==========================================
FROM base AS development
RUN go install github.com/air-verse/air@latest
EXPOSE 8080
CMD ["/go/bin/air", "-c", ".air.toml"]

# ==========================================
# Builder Stage (Compiling Assets)
# ==========================================
FROM base AS builder
# Build API binary
RUN CGO_ENABLED=1 GOOS=linux go build -o api ./cmd/api
# Build Worker binary
RUN CGO_ENABLED=1 GOOS=linux go build -o worker ./cmd/worker
# Build MockBank binary
RUN CGO_ENABLED=1 GOOS=linux go build -o mockbank ./cmd/mockbank

# ==========================================
# Production Stage
# ==========================================
FROM alpine:3.19 AS production
# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
# Copy binaries from builder
COPY --from=builder /app/api .
COPY --from=builder /app/worker .
COPY --from=builder /app/mockbank .
# Copy configuration files (if exist)
COPY config ./config
# Create non-root user for security
RUN adduser -D -u 1000 appuser
USER appuser
# Expose ports
EXPOSE 8080
EXPOSE 8081
# Default command is the API
CMD ["./api"]
