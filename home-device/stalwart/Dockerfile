# Stalwart Mail Server
# Official multi-arch image (supports arm64, amd64)
# Documentation: https://stalw.art/

FROM stalwartlabs/stalwart:0.15.4

# Copy custom configuration
COPY config.toml /opt/stalwart-mail/etc/config.toml

# Expose mail server ports
EXPOSE 25 587 993

# Use default CMD from base image (stalwart-mail binary)
