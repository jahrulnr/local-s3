# Use the official Go image as the base image
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o locals3 .

# Use a minimal base image for the final image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/locals3 .

# Create data directory
RUN mkdir -p /data

# Expose the port
EXPOSE 3000

# Set environment variables
ENV PORT=3000
ENV DATA_DIR=/data
ENV ACCESS_KEY=test
ENV SECRET_KEY=test123456789
ENV REGION=us-east-1
ENV LOG_LEVEL=info
ENV BASE_DOMAIN=localhost

# Run the application
CMD ["./locals3"]
