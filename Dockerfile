# syntax=docker/dockerfile:1

# ── Build stage ─────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /src

# CGO build deps for github.com/mattn/go-sqlite3
RUN apk add --no-cache gcc musl-dev

# Copy module files first for better layer caching
COPY go.mod go.sum ./

# Download deps once and cache them
RUN go mod download

# Copy application source
COPY cmd/ cmd/
COPY internal/ internal/
COPY templates/ templates/
COPY static/ static/

# Build release binary
RUN CGO_ENABLED=1 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/checklist ./cmd/main.go


# ── Runtime stage ───────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# Required at runtime:
# - ca-certificates for outbound TLS if needed
# - sqlite-libs for go-sqlite3 dynamic runtime dependency
# - tzdata for local time support
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

# Copy binary and runtime assets
COPY --from=builder /out/checklist /app/checklist
COPY templates/ /app/templates/
COPY static/ /app/static/

# Create writable data directory
RUN mkdir -p /data

EXPOSE 8080

ENV DB_PATH=/data/checklist.db
ENV PORT=8080

CMD ["/app/checklist"]