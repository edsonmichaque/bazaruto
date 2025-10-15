#!/bin/bash
# Post-removal script for Bazaruto

set -e

# Remove systemd service file
if [ -f /etc/systemd/system/bazaruto.service ]; then
    rm -f /etc/systemd/system/bazaruto.service
    systemctl daemon-reload
fi

# Remove the binary
if [ -f /usr/bin/bazarutod ]; then
    rm -f /usr/bin/bazarutod
fi

# Optionally remove user and directories (commented out for safety)
# Uncomment the following lines if you want to remove everything
# userdel bazaruto
# rm -rf /var/lib/bazaruto
# rm -rf /var/log/bazaruto
# rm -rf /var/cache/bazaruto

echo "Bazaruto post-removal completed successfully"
