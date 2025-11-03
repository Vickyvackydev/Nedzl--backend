# Use official golang image
FROM golang:1.24.5 AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o go-crud-api .

# Final stage - use minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/go-crud-api .

# Expose port (Railway will set PORT env var)
EXPOSE 8000

# Run the executable
CMD ["./go-crud-api"]