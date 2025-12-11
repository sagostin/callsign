# API Dockerfile for CallSign
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o callsign-api .

# -----------------------------------
# Runtime image
# -----------------------------------
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/callsign-api .

# Create directories for storage
RUN mkdir -p /var/lib/callsign/recordings \
    /var/lib/callsign/voicemail \
    /var/lib/callsign/fax \
    /var/lib/callsign/audio

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the API
CMD ["./callsign-api"]
