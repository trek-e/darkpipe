# DarkPipe Profile Server Dockerfile
# Multi-stage build for profile generation service

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimization flags
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /profile-server ./cmd/profile-server/

# Stage 2: Runtime
FROM alpine:3.21

# Add ca-certificates for HTTPS requests and create nonroot user
RUN apk add --no-cache ca-certificates wget && \
    adduser -D -u 10001 nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /profile-server /app/profile-server

# Copy templates and static assets
COPY --from=builder /build/cmd/profile-server/templates /app/templates
COPY --from=builder /build/cmd/profile-server/static /app/static

# OCI labels
LABEL org.opencontainers.image.source="https://github.com/trek-e/darkpipe"
LABEL org.opencontainers.image.description="DarkPipe Profile Server - Device onboarding and profile generation"
LABEL org.opencontainers.image.licenses="AGPL-3.0"

# Run as nonroot user
USER nonroot

# Expose profile server port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8090/health || exit 1

# Run the profile server
ENTRYPOINT ["/app/profile-server"]
