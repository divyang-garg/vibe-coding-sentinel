# Sentinel Hub API - Production Dockerfile
# Complies with CODING_STANDARDS.md: Production deployment standards

# Build stage
FROM golang:1.21-alpine AS builder

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

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty)" \
    -o sentinel-hub-api ./cmd/sentinel

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S sentinel && \
    adduser -u 1001 -S sentinel -G sentinel

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sentinel-hub-api .

# Change ownership to non-root user
RUN chown sentinel:sentinel sentinel-hub-api

# Switch to non-root user
USER sentinel

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./sentinel-hub-api"]