# Stalwart Mail Server
# Official multi-arch image (supports arm64, amd64)
# Documentation: https://stalw.art/

# Build arguments for OCI labels
ARG VERSION=dev

FROM stalwartlabs/stalwart:0.15.4

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

# Expose mail server ports
EXPOSE 25 587 993 8080

# Use entrypoint wrapper instead of direct stalwart-mail binary
ENTRYPOINT ["/entrypoint-wrapper.sh"]
