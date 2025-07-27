# 🦔 La Ninna - Multi-stage Docker Build

# Build stage
FROM golang:1.19-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git sqlite-dev gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install Swagger CLI and generate docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN ~/go/bin/swag init

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o laninna-app .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/laninna-app .

# Copy static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/fonts ./fonts
COPY --from=builder /app/docs ./docs

# Create data directory
RUN mkdir -p /data

# Set environment variables
ENV PORT=8080
ENV GIN_MODE=release
ENV DB_PATH=/data/laninna.db
ENV AUTO_MIGRATE=true

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./laninna-app"]