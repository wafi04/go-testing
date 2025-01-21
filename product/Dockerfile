# Build stage
FROM golang:1.22-alpine AS builder

# Install required packages
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# First, copy all workspace files
COPY go.work go.work.sum ./

# Copy all go.mod and go.sum files
COPY auth/go.mod auth/go.sum ./auth/
COPY category/go.mod category/go.sum ./category/
COPY common/go.mod common/go.sum ./common/
COPY product/go.mod product/go.sum ./product/
COPY gateway/go.mod gateway/go.sum ./gateway/

# Copy all service directories
COPY auth/ ./auth/
COPY category/ ./category/
COPY common/ ./common/
COPY product/ ./product/
COPY gateway/ ./gateway/

# Set work directory to the specific service
WORKDIR /app/product

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o product-service .

# Final stage
FROM alpine:3.19

# Install necessary runtime packages
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/product/product-service .

# Use non-root user
USER appuser

EXPOSE 50053

CMD ["./product-service"]