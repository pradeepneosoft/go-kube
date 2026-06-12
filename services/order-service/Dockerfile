# Stage 1: build the binary
FROM golang:1.22-alpine AS builder
# Use the official Go Alpine image as the build environment.

WORKDIR /src
# Set the working directory inside the build container.

# Copy only dependency files first for better caching.
COPY go.mod go.sum ./
# Copy Go module files so dependencies can be downloaded before the full source tree.
RUN go mod download
# Download Go modules to use the Docker build cache efficiently.

# Copy the full project source into the build stage.
COPY . .
# Build the Linux binary for the order service and place it at /out/order-service.
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/order-service ./services/order-service/cmd

# Stage 2: small runtime image
FROM alpine:3.20
# Start from a tiny Linux base image to keep the final container small.

RUN apk add --no-cache ca-certificates
# Install trusted CA certificates for secure network connections.

WORKDIR /app
# Set the runtime working directory.
COPY --from=builder /out/order-service /app/order-service
# Copy the compiled binary from the builder stage into the final image.
COPY services/order-service/migrations /app/services/order-service/migrations
# Copy the SQL migration files needed at startup.

EXPOSE 50052
# Tell Docker that the service listens on port 50052.

ENTRYPOINT ["/app/order-service"]
# Run the compiled binary when the container starts.
