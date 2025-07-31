# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static'" \
    -o forum \
    main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite

# Create non-root user for security
RUN addgroup -g 1001 -S forum && \
    adduser -S forum -u 1001 -G forum

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/forum ./forum

# Copy necessary files
COPY --from=builder /app/database/schema.sql ./database/schema.sql
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Create data directory for database
RUN mkdir -p /app/data && \
    chown -R forum:forum /app

# Switch to non-root user
USER forum

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the application
CMD ["./forum"] 