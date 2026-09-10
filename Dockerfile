# --- Build stage ---
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Install deps first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o waka-dispatch ./cmd/api

# --- Run stage ---
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates so HTTPS to Neon works
RUN apk add --no-cache ca-certificates

# Copy binary and frontend
COPY --from=builder /app/waka-dispatch .
COPY --from=builder /app/web ./web

# Render provides PORT env var; app already reads it
EXPOSE 8080

CMD ["./waka-dispatch"]