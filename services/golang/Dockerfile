# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build server and worker binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/worker ./cmd/worker

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/server /app/server
COPY --from=builder /build/worker /app/worker

EXPOSE 8080

CMD ["/app/server"]
