# Stage 1: Build the Go application
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . ./

# Define build arguments for OS and architecture
ARG GOOS=linux
ARG GOARCH=amd64

# Build the Go application with static linking
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o todo ./cmd/todo/main.go

# Stage 2: Create a minimal image to run the application using Distroless
FROM gcr.io/distroless/static-debian12

# Set the working directory inside the container
WORKDIR /app

# Copy the statically linked binary from the builder stage
COPY --from=builder /app/todo .

# Copy the migrations directory into the container
COPY --from=builder /app/migrations ./migrations

# Explicitly disable Docker health checks
HEALTHCHECK NONE

# Run the application
CMD ["/app/todo"]
