FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required dependencies
RUN apk add --no-cache git openssh-client

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o justvpn ./src

# Create a minimal image for running the application
FROM alpine:3.18

WORKDIR /app

# Install required runtime dependencies
RUN apk add --no-cache ca-certificates openssh-client

# Copy the binary from the builder stage
COPY --from=builder /app/justvpn /app/justvpn

# Copy necessary files
COPY src/users.json /app/src/users.json
COPY iac /app/iac

# Create a non-root user to run the application
RUN adduser -D -u 1000 appuser
RUN chown -R appuser:appuser /app
USER appuser

# Set environment variables
ENV TERRAFORM_WORKING_DIR=/app/src
ENV IAC_DIR_PATH=/app/iac
ENV USERS_FILE_PATH=/app/src/users.json

# Expose the API port
EXPOSE 8081

# Run the application
CMD ["/app/justvpn"]
