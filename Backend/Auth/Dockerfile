# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install git for dependency fetching
RUN apk add --no-cache git

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of your source code
COPY . .

# Build the Go binary statically for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service main.go

# Stage 2: Create a minimal image with the binary
FROM alpine:latest

# Install certificates for HTTPS connections (required for MongoDB Atlas)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the compiled binary from builder stage
COPY --from=builder /app/auth-service .

# Expose the port your app listens on (default 8080)
EXPOSE 8080

# Command to run the binary
CMD ["./auth-service"]
