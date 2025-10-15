#!/bin/bash
# Pre-installation script for Bazaruto

set -e

# Create bazaruto user if it doesn't exist
if ! id -u bazaruto >/dev/null 2>&1; then
    useradd --system --shell /bin/false --home-dir /var/lib/bazaruto --create-home bazaruto
fi

# Create necessary directories
mkdir -p /var/lib/bazaruto
mkdir -p /var/log/bazaruto
mkdir -p /var/cache/bazaruto
mkdir -p /etc/bazaruto

# Set ownership
chown -R bazaruto:bazaruto /var/lib/bazaruto
chown -R bazaruto:bazaruto /var/log/bazaruto
chown -R bazaruto:bazaruto /var/cache/bazaruto

# Set permissions
chmod 755 /var/lib/bazaruto
chmod 755 /var/log/bazaruto
chmod 755 /var/cache/bazaruto
chmod 755 /etc/bazaruto

echo "Bazaruto pre-installation completed successfully"
