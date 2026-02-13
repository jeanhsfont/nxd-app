# Build stage
FROM golang:1.21-bullseye AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy source code
COPY core/ ./core/
COPY api/ ./api/
COPY data/ ./data/
COPY services/ ./services/
COPY web/ ./web/
COPY main.go ./

# Build
RUN CGO_ENABLED=1 go build -o hub_server .

# Runtime stage - Debian para compatibilidade com sqlite
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary and web files
COPY --from=builder /build/hub_server .
COPY --from=builder /build/web ./web

# Create directories
RUN mkdir -p /app/data /app/logs

EXPOSE 8080

CMD ["./hub_server"]
