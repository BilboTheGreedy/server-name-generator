# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server-name-generator ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install necessary certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/server-name-generator .

# Copy static files and .env
COPY static/ ./static/
COPY .env .

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./server-name-generator"]