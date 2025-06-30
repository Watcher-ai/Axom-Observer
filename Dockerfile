# Multi-stage build for production-ready observer
FROM golang:1.24.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the observer
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o observer ./main.go

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 -S observer && \
    adduser -u 1001 -S observer -G observer

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/observer .

# Copy any additional files needed
COPY --from=builder /app/certs ./certs

# Create necessary directories
RUN mkdir -p /app/logs /app/certs && \
    chown -R observer:observer /app

# Switch to non-root user
USER observer

# Expose ports
EXPOSE 8888 8443 2112

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:2112/metrics || exit 1

# Default command
ENTRYPOINT ["./observer"]