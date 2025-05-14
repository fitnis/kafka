FROM golang:1.24-alpine as builder
WORKDIR /app

# Copy the entire project directory
COPY . .

# Change to the service directory
WORKDIR /app/api-gateway

# Build
RUN go mod tidy
RUN go build -o main .

# Final stage for a smaller image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app/

# Copy the binary from builder
COPY --from=builder /app/api-gateway/main .

# Run
CMD ["./main"]
