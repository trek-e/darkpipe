# Stalwart Mail Server
# Official multi-arch image (supports arm64, amd64)
# Documentation: https://stalw.art/

# Build arguments for OCI labels
ARG VERSION=dev

FROM stalwartlabs/stalwart:v0.15

# Copy custom configuration
COPY home-device/stalwart/config.toml /opt/stalwart-mail/etc/config.toml

# Copy entrypoint wrapper for setup detection and secrets support
COPY home-device/stalwart/entrypoint-wrapper.sh /entrypoint-wrapper.sh
RUN chmod +x /entrypoint-wrapper.sh

# OCI labels for container metadata
LABEL org.opencontainers.image.source=https://github.com/trek-e/darkpipe
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.description="DarkPipe Home Device - Stalwart Mail Server"
LABEL org.opencontainers.image.licenses=AGPL-3.0

# Install curl for healthcheck (Debian-based image)
RUN apt-get update && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/*

# Expose mail server ports
EXPOSE 25 587 993 8080

# Health check — verify HTTP management interface is responding
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl --silent --fail http://localhost:8080/ || exit 1

# Use entrypoint wrapper instead of direct stalwart-mail binary
ENTRYPOINT ["/entrypoint-wrapper.sh"]
