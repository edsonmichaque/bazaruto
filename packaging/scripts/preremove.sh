#!/bin/bash
# Pre-removal script for Bazaruto

set -e

# Stop the service if it's running
if systemctl is-active --quiet bazaruto; then
    echo "Stopping Bazaruto service..."
    systemctl stop bazaruto
fi

# Disable the service
if systemctl is-enabled --quiet bazaruto; then
    echo "Disabling Bazaruto service..."
    systemctl disable bazaruto
fi

echo "Bazaruto pre-removal completed successfully"
