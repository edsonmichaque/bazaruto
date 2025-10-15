#!/bin/bash
# Post-installation script for Bazaruto

set -e

# Set ownership of the binary
chown root:root /usr/bin/bazarutod
chmod 755 /usr/bin/bazarutod

# Set ownership of config file
if [ -f /etc/bazaruto/config.yaml ]; then
    chown root:root /etc/bazaruto/config.yaml
    chmod 644 /etc/bazaruto/config.yaml
fi

# Create systemd service file
cat > /etc/systemd/system/bazaruto.service << 'EOF'
[Unit]
Description=Bazaruto Backend Service
Documentation=https://github.com/edsonmichaque/bazaruto
After=network.target

[Service]
Type=simple
User=bazaruto
Group=bazaruto
ExecStart=/usr/bin/bazarutod serve
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=bazaruto

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/bazaruto /var/log/bazaruto /var/cache/bazaruto

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

echo "Bazaruto post-installation completed successfully"
echo "To start the service, run: systemctl start bazaruto"
echo "To enable the service, run: systemctl enable bazaruto"
