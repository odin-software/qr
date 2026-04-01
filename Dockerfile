# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o qr .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/qr .

# Copy static assets and templates
COPY --from=builder /build/templates ./templates
COPY --from=builder /build/static ./static

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create static directory for generated QR codes
RUN mkdir -p /app/static && chown -R appuser:appgroup /app

USER appuser

# Expose application port
EXPOSE 7003

# Set default environment variables
ENV PORT=7003
ENV ENVIRONMENT=production

# Run the application
CMD ["./qr"]
