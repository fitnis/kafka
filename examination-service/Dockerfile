FROM golang:1.24-alpine as builder

# Install required dependencies for CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy the entire project directory
COPY . .

# Change to the service directory
WORKDIR /app/examination-service

# Enable CGO and build
ENV CGO_ENABLED=1
RUN go mod tidy
RUN go build -o main .

# Final stage for a smaller image
FROM alpine:latest
RUN apk --no-cache add ca-certificates gcc musl-dev
WORKDIR /app/

# Copy the binary from builder
COPY --from=builder /app/examination-service/main .
# Copy the shared database file if needed
COPY --from=builder /app/fitnis.db /app/fitnis.db

# Run
CMD ["./main"]
