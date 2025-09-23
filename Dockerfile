# Build stage
FROM golang:1.24.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata && \
	addgroup -S appgroup && \
	adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy web assets and articles
COPY --from=builder /app/web ./web
COPY --from=builder /app/articles ./articles

# Create data directory with proper permissions
RUN mkdir -p data logs && \
	chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 7777

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
	CMD wget --no-verbose --tries=1 --spider http://localhost:7777/ || exit 1

# Environment variables
ENV PORT=7777
ENV DATABASE_PATH=/app/data/admin.db
ENV ARTICLES_DIR=/app/articles
ENV TEMPLATE_DIR=/app/web/templates
ENV STATIC_DIR=/app/web/static
ENV DEBUG_MODE=false

# Run the application
CMD ["./main"]
