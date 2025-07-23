# --------- Stage 1: Build the Go binary ---------
FROM golang:1.24-alpine AS builder

# Install Git (for go mod) and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -o voting-service main.go

# --------- Stage 2: Run with minimal image ---------
FROM alpine:latest

# Install certs required for HTTPS (MongoDB Atlas, etc.)
RUN apk add --no-cache ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/voting-service .

# Expose the app port
EXPOSE 8082

# Run the binary
CMD ["./voting-service"]
