# WTE

[![CI](https://github.com/wtepcorp/WTE/actions/workflows/ci.yml/badge.svg)](https://github.com/wtepcorp/WTE/actions/workflows/ci.yml)
[![Release](https://github.com/wtepcorp/WTE/actions/workflows/release.yml/badge.svg)](https://github.com/wtepcorp/WTE/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**Window to Europe** — CLI tool for managing proxy infrastructure based on GOST. Easily install, configure, and manage HTTP, HTTPS, and Shadowsocks proxy servers on Linux.

## Features

- **HTTP Proxy** — Classic proxy for browsers and applications
- **HTTPS Proxy** — Proxy with TLS encryption
- **Shadowsocks** — Protocol for bypassing restrictions
- **Flexible Configuration** — Enable/disable authentication, choose ports
- **Automatic Setup** — Firewall, systemd, password generation
- **Easy Management** — start, stop, restart, status, logs

---

## Quick Install

```bash
curl -sfL https://raw.githubusercontent.com/wtepcorp/WTE/main/install.sh | sudo bash
```

## Manual Installation

### Option 1: Download Binary

```bash
# Download latest version
wget https://github.com/wtepcorp/WTE/releases/latest/download/wte-linux-amd64.tar.gz

# Extract
tar -xzf wte-linux-amd64.tar.gz

# Install
sudo mv wte-linux-amd64 /usr/local/bin/wte
sudo chmod +x /usr/local/bin/wte

# Verify
wte version
```

### Option 2: Build from Source

```bash
# Install Go (if not installed)
sudo apt update && sudo apt install -y golang-go

# Clone repository
git clone https://github.com/wtepcorp/WTE.git
cd WTE

# Build and install
make build
sudo make install

# Verify
wte version
```

---

## Quick Start

### Install Proxy Server

```bash
# Basic installation (HTTP + Shadowsocks with auto-generated passwords)
sudo wte install

# HTTP proxy only (no Shadowsocks)
sudo wte install --ss-enabled=false

# Installation without authentication
sudo wte install --http-no-auth

# Custom settings
sudo wte install \
    --http-port 3128 \
    --http-user admin \
    --http-pass mypassword \
    --ss-port 8388
```

Connection credentials will be displayed after installation.

### Service Management

```bash
# Check status
sudo wte status

# Stop
sudo wte stop

# Start
sudo wte start

# Restart
sudo wte restart

# View logs
sudo wte logs

# Follow logs in real-time
sudo wte logs -f
```

### View Credentials

```bash
# Show all connection details
sudo wte credentials

# Show Shadowsocks URI only (for import)
sudo wte credentials --uri

# Regenerate passwords
sudo wte credentials --regenerate
```

### Configuration Management

```bash
# Show current configuration
wte config show

# Change HTTP proxy port
sudo wte config set http.port 3128

# Disable authentication
sudo wte config set http.auth.enabled false

# Enable Shadowsocks
sudo wte config set shadowsocks.enabled true

# Apply changes (regenerate config and restart)
sudo wte config apply

# Open config in editor
sudo wte config edit

# Reset to defaults
sudo wte config reset
```

### Update WTE

```bash
# Check for updates
sudo wte update --check

# Update to latest version
sudo wte update

# Force reinstall
sudo wte update --force
```

### Uninstall

```bash
# Complete uninstall
sudo wte uninstall

# Uninstall without confirmation
sudo wte uninstall --force

# Uninstall but keep credentials file
sudo wte uninstall --keep-creds
```

---

## Installation Options

| Flag | Description | Default |
|------|-------------|---------|
| `--http-port` | HTTP proxy port | 8080 |
| `--http-user` | Username | proxyuser |
| `--http-pass` | Password (auto-generated if empty) | — |
| `--http-no-auth` | Disable authentication | false |
| `--ss-enabled` | Enable Shadowsocks | true |
| `--ss-port` | Shadowsocks port | 9500 |
| `--ss-password` | SS password (auto-generated if empty) | — |
| `--ss-method` | Encryption method | aes-128-gcm |
| `--https-enabled` | Enable HTTPS proxy | false |
| `--https-port` | HTTPS proxy port | 8443 |
| `--skip-firewall` | Don't configure firewall | false |
| `--gost-version` | GOST version | 3.0.0-rc10 |

---

## Connecting to Proxy

### HTTP Proxy

**Browser / System Settings:**
```
Host: <server IP>
Port: 8080
Login: proxyuser
Password: <your password>
```

**curl:**
```bash
curl -x http://proxyuser:PASSWORD@SERVER_IP:8080 https://ifconfig.me
```

**Environment Variables:**
```bash
export http_proxy="http://proxyuser:PASSWORD@SERVER_IP:8080"
export https_proxy="http://proxyuser:PASSWORD@SERVER_IP:8080"
```

### Shadowsocks

**Client Settings:**
```
Server: <server IP>
Port: 9500
Password: <your password>
Encryption: aes-128-gcm
```

**Clients:**
- **iOS:** Shadowrocket, Surge, Quantumult
- **Android:** Shadowsocks, v2rayNG
- **Windows:** Shadowsocks-windows, v2rayN
- **macOS:** ShadowsocksX-NG, Surge
- **Linux:** shadowsocks-libev, shadowsocks-rust

**Import via URI:**
```bash
# Get URI for import
sudo wte credentials --uri
# Example: ss://YWVzLTEyOC1nY206cGFzc3dvcmQ=@1.2.3.4:9500#WTE-Proxy
```

---

## File Locations

| File | Description |
|------|-------------|
| `/usr/local/bin/gost` | GOST binary |
| `/etc/wte/config.yaml` | WTE configuration |
| `/etc/gost/config.yaml` | GOST configuration |
| `/etc/systemd/system/gost.service` | Systemd service |
| `/root/proxy-credentials.txt` | Credentials file |

---

## Usage Examples

### Example 1: Simple Proxy for Personal Use

```bash
sudo wte install
sudo wte credentials
```

### Example 2: Public Proxy Without Authentication

```bash
# Warning: not recommended for public servers!
sudo wte install --http-no-auth --ss-enabled=false
```

### Example 3: Shadowsocks Only on Non-Standard Port

```bash
sudo wte install \
    --http-port 0 \
    --ss-enabled=true \
    --ss-port 443 \
    --ss-method chacha20-ietf-poly1305
```

### Example 4: Corporate Proxy with HTTPS

```bash
sudo wte install \
    --http-port 3128 \
    --http-user corporate \
    --http-pass SecurePass123 \
    --https-enabled \
    --https-port 3129 \
    --ss-enabled=false
```

---

## Troubleshooting

### Service Not Starting

```bash
# Check status
sudo systemctl status gost

# View logs
sudo journalctl -u gost -n 50

# Check configuration
cat /etc/gost/config.yaml
```

### Port Already in Use

```bash
# Check what's using the port
sudo ss -tlnp | grep 8080

# Use different port
sudo wte config set http.port 3128
sudo wte config apply
```

### Cannot Connect from Outside

```bash
# Check firewall
sudo ufw status
# or
sudo firewall-cmd --list-all

# Manually open port (UFW)
sudo ufw allow 8080/tcp
sudo ufw allow 9500/tcp
sudo ufw allow 9500/udp
```

### Reset and Start Fresh

```bash
sudo wte uninstall --force
sudo wte install
```

---

## Global Flags

| Flag | Description |
|------|-------------|
| `-c, --config` | Config file path |
| `-v, --verbose` | Verbose output |
| `-q, --quiet` | Minimal output (errors only) |
| `--no-color` | Disable colored output |
| `-h, --help` | Show help |

---

## Requirements

- **OS:** Ubuntu 18.04+, Debian 10+, CentOS 7+, Fedora 38+, Arch Linux
- **Architecture:** x86_64 (amd64), ARM64, ARMv7
- **Privileges:** root (sudo)
- **Network:** Access to GitHub for downloading GOST

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT License - see [LICENSE](LICENSE) file for details.
