# Packaging Guide

This guide covers building and packaging the Bazaruto Insurance Platform for various operating systems and deployment scenarios.

## Build System

The project uses a Makefile for build automation and Goreleaser for release packaging.

### Makefile Targets

```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run database migrations
make migrate

# Clean build artifacts
make clean

# Install dependencies
make deps

# Format code
make fmt

# Lint code
make lint

# Generate documentation
make docs
```

## Goreleaser Configuration

The project uses Goreleaser for automated releases and packaging.

### .goreleaser.yaml

```yaml
# .goreleaser.yaml
project_name: bazaruto
before:
  hooks:
    - go mod download
    - go mod verify
builds:
  - id: bazarutod
    binary: bazarutod
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
archives:
  - id: bazarutod
    builds:
      - bazarutod
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
checksums:
  name_template: 'checksums.txt'
snapcrafts:
  - name: bazarutod
    builds:
      - bazarutod
    summary: Bazaruto Insurance Platform
    description: |
      A production-grade Go backend service for insurance marketplace platform
    grade: stable
    confinement: strict
    apps:
      bazarutod:
        command: bazarutod
        plugs:
          - network
          - network-bind
          - home
          - removable-media
nfpms:
  - id: bazarutod
    builds:
      - bazarutod
    homepage: https://github.com/edsonmichaque/bazaruto
    description: Bazaruto Insurance Platform
    license: MIT
    maintainer: Edson Michaque <edson@example.com>
    vendor: Edson Michaque
    formats:
      - deb
      - rpm
    dependencies:
      - postgresql-client
      - redis-tools
    contents:
      - src: README.md
        dst: /usr/share/doc/bazarutod/README.md
      - src: LICENSE
        dst: /usr/share/doc/bazarutod/LICENSE
      - src: config.yaml.example
        dst: /etc/bazarutod/config.yaml.example
    config_files:
      - /etc/bazarutod/config.yaml
    scripts:
      preinstall: scripts/preinstall.sh
      postinstall: scripts/postinstall.sh
      preremove: scripts/preremove.sh
      postremove: scripts/postremove.sh
dockers:
  - id: bazarutod
    builds:
      - bazarutod
    image_templates:
      - "bazaruto:{{ .Version }}"
      - "bazaruto:latest"
    dockerfile: Dockerfile
    use: buildx
    build_flag_templates:
      - --platform=linux/amd64
      - --platform=linux/arm64
```

## Package Types

### 1. Binary Releases

Goreleaser creates cross-platform binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

### 2. Archive Packages

- **Linux/macOS**: `.tar.gz` archives
- **Windows**: `.zip` archives

### 3. System Packages

#### Debian/Ubuntu (.deb)

```bash
# Install from .deb package
sudo dpkg -i bazarutod_1.0.0_amd64.deb

# Install dependencies
sudo apt-get install -f

# Configure
sudo cp /etc/bazarutod/config.yaml.example /etc/bazarutod/config.yaml
sudo nano /etc/bazarutod/config.yaml

# Start service
sudo systemctl start bazarutod
sudo systemctl enable bazarutod
```

#### Red Hat/CentOS (.rpm)

```bash
# Install from .rpm package
sudo rpm -i bazarutod-1.0.0-1.x86_64.rpm

# Configure
sudo cp /etc/bazarutod/config.yaml.example /etc/bazarutod/config.yaml
sudo nano /etc/bazarutod/config.yaml

# Start service
sudo systemctl start bazarutod
sudo systemctl enable bazarutod
```

#### Snap Package

```bash
# Install snap package
sudo snap install bazarutod

# Configure
sudo snap set bazarutod config=/path/to/config.yaml

# Start service
sudo snap start bazarutod
```

### 4. Docker Images

#### Multi-architecture Docker Images

```bash
# Pull image
docker pull bazaruto:latest

# Run container
docker run -d \
  --name bazaruto \
  -p 8080:8080 \
  -v /path/to/config.yaml:/app/config.yaml \
  bazaruto:latest

# Run with environment variables
docker run -d \
  --name bazaruto \
  -p 8080:8080 \
  -e BAZARUTO_DB_HOST=postgres \
  -e BAZARUTO_DB_NAME=bazaruto \
  -e BAZARUTO_DB_USER=postgres \
  -e BAZARUTO_DB_PASSWORD=password \
  -e BAZARUTO_REDIS_ADDRESS=redis:6379 \
  bazaruto:latest
```

## Build Scripts

### Pre-install Script

```bash
#!/bin/bash
# scripts/preinstall.sh

# Create user and group
useradd -r -s /bin/false bazarutod || true
groupadd -r bazarutod || true

# Create directories
mkdir -p /var/lib/bazarutod
mkdir -p /var/log/bazarutod
mkdir -p /etc/bazarutod

# Set permissions
chown -R bazarutod:bazarutod /var/lib/bazarutod
chown -R bazarutod:bazarutod /var/log/bazarutod
chown -R bazarutod:bazarutod /etc/bazarutod
```

### Post-install Script

```bash
#!/bin/bash
# scripts/postinstall.sh

# Create systemd service
cat > /etc/systemd/system/bazarutod.service << EOF
[Unit]
Description=Bazaruto Insurance Platform
After=network.target

[Service]
Type=simple
User=bazarutod
Group=bazarutod
ExecStart=/usr/bin/bazarutod serve
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=bazarutod

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

# Enable service
systemctl enable bazarutod
```

### Pre-remove Script

```bash
#!/bin/bash
# scripts/preremove.sh

# Stop service
systemctl stop bazarutod || true
systemctl disable bazarutod || true
```

### Post-remove Script

```bash
#!/bin/bash
# scripts/postremove.sh

# Remove systemd service
rm -f /etc/systemd/system/bazarutod.service
systemctl daemon-reload

# Remove user and group (optional)
# userdel bazarutod || true
# groupdel bazarutod || true
```

## Docker Build

### Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bazarutod cmd/bazarutod/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/bazarutod .
COPY --from=builder /app/config.yaml.example ./config.yaml.example

EXPOSE 8080

CMD ["./bazarutod", "serve"]
```

### Multi-stage Build

```dockerfile
# Dockerfile.multi
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bazarutod cmd/bazarutod/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/bazarutod /bazarutod
COPY --from=builder /app/config.yaml.example /config.yaml.example

EXPOSE 8080

ENTRYPOINT ["/bazarutod"]
```

## Build Commands

### Local Build

```bash
# Build for current platform
make build

# Build for specific platform
GOOS=linux GOARCH=amd64 make build

# Build with version info
VERSION=1.0.0 COMMIT=$(git rev-parse HEAD) DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) make build
```

### Cross-compilation

```bash
# Build for multiple platforms
make build-all

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o bin/bazarutod-linux-amd64 cmd/bazarutod/main.go
GOOS=darwin GOARCH=amd64 go build -o bin/bazarutod-darwin-amd64 cmd/bazarutod/main.go
GOOS=windows GOARCH=amd64 go build -o bin/bazarutod-windows-amd64.exe cmd/bazarutod/main.go
```

### Docker Build

```bash
# Build Docker image
docker build -t bazaruto:latest .

# Build multi-architecture image
docker buildx build --platform linux/amd64,linux/arm64 -t bazaruto:latest .

# Build with build args
docker build --build-arg VERSION=1.0.0 -t bazaruto:1.0.0 .
```

## Release Process

### Automated Release

```bash
# Create and push tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Run Goreleaser
goreleaser release --rm-dist
```

### Manual Release

```bash
# Build all packages
goreleaser build --snapshot

# Test packages
goreleaser release --snapshot --skip-publish

# Publish release
goreleaser release
```

## Package Verification

### Binary Verification

```bash
# Verify binary
./bazarutod version

# Check binary info
file bazarutod
ldd bazarutod  # Linux
otool -L bazarutod  # macOS
```

### Package Verification

```bash
# Verify .deb package
dpkg-deb -I bazarutod_1.0.0_amd64.deb
dpkg-deb -c bazarutod_1.0.0_amd64.deb

# Verify .rpm package
rpm -qip bazarutod-1.0.0-1.x86_64.rpm
rpm -qlp bazarutod-1.0.0-1.x86_64.rpm
```

### Docker Image Verification

```bash
# Verify Docker image
docker inspect bazaruto:latest

# Test Docker image
docker run --rm bazaruto:latest version

# Check image layers
docker history bazaruto:latest
```

## Distribution

### GitHub Releases

Goreleaser automatically creates GitHub releases with:
- Binary downloads
- Archive packages
- System packages
- Docker images
- Release notes

### Package Repositories

#### Debian/Ubuntu Repository

```bash
# Create repository structure
mkdir -p repo/dists/stable/main/binary-amd64
mkdir -p repo/pool/main/b

# Add package
cp bazarutod_1.0.0_amd64.deb repo/pool/main/b/

# Generate repository metadata
cd repo
dpkg-scanpackages pool/ > dists/stable/main/binary-amd64/Packages
gzip -k dists/stable/main/binary-amd64/Packages
```

#### RPM Repository

```bash
# Create repository
mkdir -p repo/x86_64
cp bazarutod-1.0.0-1.x86_64.rpm repo/x86_64/

# Generate repository metadata
cd repo
createrepo .
```

### Snap Store

```bash
# Build snap package
snapcraft

# Test snap package
snap install --dangerous bazarutod_1.0.0_amd64.snap

# Upload to snap store
snapcraft upload bazarutod_1.0.0_amd64.snap
```

## Installation Scripts

### Linux Installation Script

```bash
#!/bin/bash
# install.sh

set -e

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VERSION=$VERSION_ID
else
    echo "Unsupported operating system"
    exit 1
fi

# Download and install based on OS
case $OS in
    ubuntu|debian)
        wget https://github.com/edsonmichaque/bazaruto/releases/latest/download/bazarutod_1.0.0_amd64.deb
        sudo dpkg -i bazarutod_1.0.0_amd64.deb
        sudo apt-get install -f
        ;;
    centos|rhel|fedora)
        wget https://github.com/edsonmichaque/bazaruto/releases/latest/download/bazarutod-1.0.0-1.x86_64.rpm
        sudo rpm -i bazarutod-1.0.0-1.x86_64.rpm
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Configure
sudo cp /etc/bazarutod/config.yaml.example /etc/bazarutod/config.yaml
echo "Please edit /etc/bazarutod/config.yaml and start the service with: sudo systemctl start bazarutod"
```

### macOS Installation Script

```bash
#!/bin/bash
# install-macos.sh

set -e

# Download binary
wget https://github.com/edsonmichaque/bazaruto/releases/latest/download/bazarutod_1.0.0_darwin_amd64.tar.gz

# Extract
tar -xzf bazarutod_1.0.0_darwin_amd64.tar.gz

# Install
sudo mv bazarutod /usr/local/bin/
sudo chmod +x /usr/local/bin/bazarutod

# Create config directory
sudo mkdir -p /usr/local/etc/bazarutod
sudo cp config.yaml.example /usr/local/etc/bazarutod/config.yaml

echo "Installation complete. Please edit /usr/local/etc/bazarutod/config.yaml"
```

### Windows Installation Script

```powershell
# install-windows.ps1

# Download binary
Invoke-WebRequest -Uri "https://github.com/edsonmichaque/bazaruto/releases/latest/download/bazarutod_1.0.0_windows_amd64.zip" -OutFile "bazarutod.zip"

# Extract
Expand-Archive -Path "bazarutod.zip" -DestinationPath "C:\Program Files\Bazaruto"

# Add to PATH
$env:PATH += ";C:\Program Files\Bazaruto"

# Create config directory
New-Item -ItemType Directory -Path "C:\Program Files\Bazaruto\config" -Force
Copy-Item "config.yaml.example" "C:\Program Files\Bazaruto\config\config.yaml"

Write-Host "Installation complete. Please edit C:\Program Files\Bazaruto\config\config.yaml"
```

## Troubleshooting

### Build Issues

1. **Go version mismatch**
   ```bash
   # Check Go version
   go version
   
   # Update Go if needed
   go install golang.org/dl/go1.22.0@latest
   go1.22.0 download
   ```

2. **Module issues**
   ```bash
   # Clean module cache
   go clean -modcache
   
   # Download modules
   go mod download
   ```

3. **Cross-compilation issues**
   ```bash
   # Enable CGO for cross-compilation
   CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build
   ```

### Package Issues

1. **Dependency issues**
   ```bash
   # Install dependencies
   sudo apt-get install -f  # Debian/Ubuntu
   sudo yum install -y postgresql-client redis-tools  # CentOS/RHEL
   ```

2. **Permission issues**
   ```bash
   # Fix permissions
   sudo chown -R bazarutod:bazarutod /var/lib/bazarutod
   sudo chmod 755 /var/lib/bazarutod
   ```

3. **Service issues**
   ```bash
   # Check service status
   sudo systemctl status bazarutod
   
   # View logs
   sudo journalctl -u bazarutod -f
   ```

For more detailed troubleshooting, refer to the [Operations Guide](ops-guide.md).


