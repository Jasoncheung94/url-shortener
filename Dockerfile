# Use the official Go image as the base for building
FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go application
RUN CGO_ENABLED=0 go build -o url-shortener ./cmd/url-shortener

# Create a small final image using the 'alpine' base for security and efficiency
FROM alpine:latest

# Install dependencies (needed for Go to serve static files)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/url-shortener /app/url-shortener

# Copy the web directory with static assets into the image
COPY --from=builder /app/web /app/web

EXPOSE 8080

CMD ["/app/url-shortener"]
