# Maddy Mail Server
# Official image (supports arm64, amd64)
# Documentation: https://maddy.email/
# Note: Maddy is currently beta - use with caution for production

# Build arguments for OCI labels
ARG VERSION=dev

FROM foxcpp/maddy:0.8.2

# Copy custom configuration
COPY home-device/maddy/maddy.conf /data/maddy.conf

# Copy alias mapping file
COPY home-device/maddy/aliases /data/aliases

# Copy entrypoint wrapper for setup detection and secrets support
COPY home-device/maddy/entrypoint-wrapper.sh /entrypoint-wrapper.sh
RUN chmod +x /entrypoint-wrapper.sh

# OCI labels for container metadata
LABEL org.opencontainers.image.source=https://github.com/trek-e/darkpipe
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.description="DarkPipe Home Device - Maddy Mail Server"
LABEL org.opencontainers.image.licenses=AGPL-3.0

# Expose mail server ports
EXPOSE 25 587 993

# Use entrypoint wrapper instead of direct maddy binary
ENTRYPOINT ["/entrypoint-wrapper.sh"]
