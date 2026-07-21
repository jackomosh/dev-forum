# --- Stage 1: Build Stage ---
FROM golang:1.22-alpine AS builder

# Install build tools required for CGO + SQLite
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build with CGO enabled
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o app ./cmd/forum

# --- Stage 2: Final Stage ---
FROM alpine:latest

# Install sqlite-libs/ca-certificates if needed at runtime
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Copy binary and web directory from builder
COPY --from=builder /app/app .
COPY --from=builder /app/web ./web

# Copy schema to image 
COPY --from=builder /app/schema.sql ./schema.sql

EXPOSE 8089

CMD ["./app"]
