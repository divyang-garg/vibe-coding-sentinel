# Sentinel Hub Document Processor
# Contains all document parsing dependencies (poppler, tesseract, etc.)

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
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o sentinel-processor .

# =============================================================================
# Production image with document processing tools
# =============================================================================
FROM ubuntu:22.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install document processing dependencies
# These are installed HERE on the server, NOT on developer machines
RUN apt-get update && apt-get install -y --no-install-recommends \
    # PDF processing
    poppler-utils \
    # OCR for images
    tesseract-ocr \
    tesseract-ocr-eng \
    tesseract-ocr-spa \
    tesseract-ocr-fra \
    tesseract-ocr-deu \
    tesseract-ocr-chi-sim \
    # Document conversion
    pandoc \
    # Spreadsheet processing
    gnumeric \
    # LibreOffice for complex documents (optional, large)
    # libreoffice-writer \
    # Archive handling
    unzip \
    # Utilities
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/sentinel-processor /usr/local/bin/

# Create data directory
RUN mkdir -p /data/documents && chmod 755 /data/documents

# Create non-root user
RUN useradd -m -s /bin/bash appuser && \
    chown -R appuser:appuser /data
USER appuser

VOLUME ["/data/documents"]

CMD ["sentinel-processor"]

