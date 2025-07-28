# 🦔 La Ninna - Multi-stage Docker Build

# Build stage
FROM golang:1.22-alpine3.19 AS builder

# Install build dependencies
RUN apk add --no-cache git sqlite-dev gcc musl-dev

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Install Swagger CLI and Air for hot reloading
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go install github.com/cosmtrek/air@latest

# Copy source code
COPY . .

# Generate Swagger docs
RUN ~/go/bin/swag init

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o laninna-app .

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create app directories
WORKDIR /app
RUN mkdir -p /data && chown -R appuser:appgroup /data

# Copy binary from builder stage
COPY --from=builder /app/laninna-app .

# Copy static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/fonts ./fonts
COPY --from=builder /app/docs ./docs

# Copy Air configuration for development
COPY --from=builder /app/.air.toml ./.air.toml

# Set environment variables
ENV PORT=8080
ENV GIN_MODE=release
ENV DB_PATH=/data/laninna.db
ENV AUTO_MIGRATE=true
ENV FONTS_PATH=./fonts

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Switch to non-root user
USER appuser

# Run the application
CMD ["./laninna-app"]