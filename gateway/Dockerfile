FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

# Copy seluruh project
COPY . .

# Set working directory ke auth service
WORKDIR /app/gateway

# Download dependencies
RUN go mod download


RUN go get -v github.com/jackc/pgx/v5/stdlib \
    && go get -v github.com/google/uuid \
    && go get -v google.golang.org/grpc \
    && go get -v google.golang.org/protobuf/types/known/timestamppb \
    && go get github.com/wafi04/go-testing/common \
    && go get github.com/wafi04/go-testing/auth \
    && go get github.com/wafi04/go-testing/category \
    && go get github.com/wafi04/go-testing/product

# Tidy up modules
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o main .

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy built binary (Fix the path here)
COPY --from=builder /app/gateway/main .

# Set user
USER appuser

EXPOSE 50051

CMD ["./main"]