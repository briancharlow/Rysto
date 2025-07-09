# Dockerfile for Story Service

# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install git
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o story-service main.go

# Stage 2: Minimal image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/story-service .
EXPOSE 8081
CMD ["./story-service"]