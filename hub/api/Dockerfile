# Sentinel Hub API Server
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o sentinel-hub-api .

# =============================================================================
# Production image
# =============================================================================
FROM alpine:3.19

RUN apk add --no-cache \
    ca-certificates \
    curl \
    poppler-utils \
    fontconfig \
    ttf-dejavu \
    libzip \
    unzip \
    tesseract-ocr \
    tesseract-ocr-data-eng

WORKDIR /app

COPY --from=builder /app/sentinel-hub-api /usr/local/bin/

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["sentinel-hub-api"]









