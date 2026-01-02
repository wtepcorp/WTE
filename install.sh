#!/bin/bash
#
# WTE - Installation Script
#
# Usage:
#   curl -sfL https://raw.githubusercontent.com/wtepcorp/WTE/main/install.sh | sudo bash
#
# Or with specific version:
#   curl -sfL https://raw.githubusercontent.com/wtepcorp/WTE/main/install.sh | sudo bash -s -- v1.0.0
#

set -e

# Configuration
REPO="wtepcorp/WTE"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="wte"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Check if running as root
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        error "This script must be run as root. Use: sudo bash install.sh"
    fi
}

# Detect OS
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
    else
        error "Cannot detect operating system"
    fi

    case $OS in
        ubuntu|debian|centos|rhel|rocky|almalinux|fedora|arch)
            success "Detected OS: $OS"
            ;;
        *)
            warning "OS '$OS' is not officially supported, proceeding anyway..."
            ;;
    esac
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac
    success "Detected architecture: $ARCH"
}

# Get latest version from GitHub
get_latest_version() {
    info "Fetching latest version..."

    LATEST_VERSION=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" | \
        grep '"tag_name"' | \
        sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        error "Could not fetch latest version. Check your internet connection."
    fi

    success "Latest version: $LATEST_VERSION"
}

# Download and install
install_wte() {
    local version=$1
    local arch=$2

    local download_url="https://github.com/$REPO/releases/download/$version/wte-linux-$arch.tar.gz"
    local temp_dir=$(mktemp -d)

    info "Downloading WTE $version for linux/$arch..."
    info "URL: $download_url"

    if ! curl -sfL "$download_url" -o "$temp_dir/wte.tar.gz"; then
        rm -rf "$temp_dir"
        error "Download failed. Check if the release exists."
    fi

    success "Download completed"

    info "Extracting..."
    tar -xzf "$temp_dir/wte.tar.gz" -C "$temp_dir"

    # Find the binary
    local binary=$(find "$temp_dir" -name "wte*" -type f -executable | head -1)
    if [ -z "$binary" ]; then
        binary=$(find "$temp_dir" -name "wte*" -type f | head -1)
    fi

    if [ -z "$binary" ]; then
        rm -rf "$temp_dir"
        error "Binary not found in archive"
    fi

    info "Installing to $INSTALL_DIR/$BINARY_NAME..."

    # Backup existing binary if exists
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        mv "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME.backup"
        info "Backed up existing binary"
    fi

    mv "$binary" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    rm -rf "$temp_dir"

    success "Installation completed"
}

# Verify installation
verify_installation() {
    info "Verifying installation..."

    if ! command -v wte &> /dev/null; then
        error "Installation failed. 'wte' command not found."
    fi

    local installed_version=$(wte version 2>/dev/null | head -1)
    success "Installed: $installed_version"
}

# Main
main() {
    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║              WTE - Window to Europe                       ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""

    check_root
    detect_os
    detect_arch

    # Use provided version or get latest
    VERSION=${1:-""}
    if [ -z "$VERSION" ]; then
        get_latest_version
        VERSION=$LATEST_VERSION
    else
        info "Using specified version: $VERSION"
    fi

    install_wte "$VERSION" "$ARCH"
    verify_installation

    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║           ✓ Installation Successful!                      ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""
    echo "Quick start:"
    echo "  sudo wte install    # Install proxy server"
    echo "  sudo wte status     # Check status"
    echo "  wte --help          # Show all commands"
    echo ""
}

main "$@"
