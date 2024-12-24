# Build stage
FROM golang:1.23 AS builder
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Run the application
CMD ["./main"]