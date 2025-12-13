# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY seed.go ./
COPY data/ ./data/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /seed seed.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary and data files
COPY --from=builder /seed /app/seed
COPY --from=builder /app/data /app/data

# Run the binary
CMD ["/app/seed"]
