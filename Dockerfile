# Stage 1: Build the Go application
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/todo/main.go

# Stage 2: Create a minimal image to run the application
FROM alpine:latest

# Install certificates (if your application uses HTTPS)
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the statically linked binary from the builder stage
COPY --from=builder /app/main .

# Copy the migrations directory into the container
COPY --from=builder /app/migrations ./migrations

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]
