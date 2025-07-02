# K_CMS Application Dockerfile
FROM golang:1.22-alpine AS builder

# Install gcc and other build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main .

# Use alpine for the final image
FROM alpine:latest

# Install ca-certificates for HTTPS requests and gcc for runtime
RUN apk --no-cache add ca-certificates gcc musl-dev

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
