# Use a base image with Go and Chrome
FROM golang:1.21-bullseye

# Install Chrome and dependencies
RUN apt-get update && apt-get install -y \
    chromium \
    chromium-driver \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code with new structure
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY configs/ ./configs/

# Create directories for output and logs
RUN mkdir -p output logs temp_builds

# Set environment variables
ENV CHROME_BIN=/usr/bin/chromium
ENV CHROME_PATH=/usr/lib/chromium/
ENV GOTMPDIR=/app/temp_builds
ENV CONFIG_PATH=/app/configs/config.yaml

# Build the application
RUN go build -o webscraper ./cmd/main.go

# Create entrypoint script
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"] 