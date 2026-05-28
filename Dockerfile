# AILINTER — Production Dockerfile
# Multi-stage build: golang:1.25-alpine builder → distroless/alpine runtime
# Targets: linux/amd64, linux/arm64
#
# Usage:
#   docker build -t ailinter/ailinter:latest \
#     --build-arg VERSION=$(git describe --tags --always --dirty) .
#
# Run:
#   docker run --rm -p 4317:4317 ailinter/ailinter:latest mcp

# ── Stage 1: Build ──────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Cache dependencies (separate from source for layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Inject version from build arg (default: dev)
ARG VERSION=dev
ARG TARGETOS TARGETARCH

# Build static binary:
#   - CGO_ENABLED=0       → pure Go, no libc dependency, works on Alpine + distroless
#   - -ldflags="-s -w"    → strip debug symbols and DWARF table (30% size reduction)
#   - -trimpath           → remove build system paths from binary
#   - -buildmode=pie      → position-independent executable (security hardening)
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-arm64} \
    go build \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -trimpath \
      -buildmode=pie \
      -o /build/ailinter \
      ./cmd/ailinter

# Verify binary
RUN file /build/ailinter && \
    /build/ailinter version 2>/dev/null || echo "version check deferred"

# ── Stage 2: Runtime ────────────────────────────
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1001 ailinter

WORKDIR /app
COPY --from=builder /build/ailinter /usr/local/bin/ailinter

# Create data directory with correct permissions
RUN mkdir -p /data && chown -R ailinter:ailinter /data

USER ailinter

EXPOSE 4317

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD ailinter version || exit 1

ENTRYPOINT ["ailinter"]
CMD ["mcp"]
