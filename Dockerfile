# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache --update \
    build-base \
    sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o forum ./cmd/forum

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache --update \
    ca-certificates \
    sqlite \
    tzdata

WORKDIR /app

COPY --from=builder /app/forum .
COPY --from=builder /app/web ./web
COPY --from=builder /app/schema.sql ./schema.sql

RUN mkdir -p /app/data

EXPOSE 8089

ENV FORUM_PORT=8089 \
    FORUM_DB_PATH=/app/data/forum.db

CMD ["./forum"]
