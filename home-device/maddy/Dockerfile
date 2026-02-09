# Maddy Mail Server
# Official image (supports arm64, amd64)
# Documentation: https://maddy.email/
# Note: Maddy is currently beta - use with caution for production

FROM foxcpp/maddy:0.8.2

# Copy custom configuration
COPY maddy.conf /data/maddy.conf

# Copy alias mapping file
COPY aliases /data/aliases

# Expose mail server ports
EXPOSE 25 587 993

# Use default CMD from base image (maddy binary)
