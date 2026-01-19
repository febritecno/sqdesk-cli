#!/bin/bash

# SQDesk Installer Script
# Usage: curl -fsSL https://raw.githubusercontent.com/febritecno/sqdesk-cli/main/install.sh | bash

set -e

REPO="febritecno/sqdesk-cli"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="sqdesk"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "  ███████╗ ██████╗ ██████╗ ███████╗███████╗██╗  ██╗"
echo "  ██╔════╝██╔═══██╗██╔══██╗██╔════╝██╔════╝██║ ██╔╝"
echo "  ███████╗██║   ██║██║  ██║█████╗  ███████╗█████╔╝ "
echo "  ╚════██║██║▄▄ ██║██║  ██║██╔══╝  ╚════██║██╔═██╗ "
echo "  ███████║╚██████╔╝██████╔╝███████╗███████║██║  ██╗"
echo "  ╚══════╝ ╚══▀▀═╝ ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝"
echo -e "${NC}"
echo "  Modern SQL Client for Terminal"
echo ""

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS="linux";;
        Darwin*)    OS="darwin";;
        CYGWIN*|MINGW*|MSYS*) OS="windows";;
        *)          OS="unknown";;
    esac
    echo $OS
}

# Detect Architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   ARCH="amd64";;
        arm64|aarch64)  ARCH="arm64";;
        armv7l)         ARCH="arm";;
        i386|i686)      ARCH="386";;
        *)              ARCH="unknown";;
    esac
    echo $ARCH
}

OS=$(detect_os)
ARCH=$(detect_arch)

echo -e "${YELLOW}Detected:${NC} ${OS}/${ARCH}"

if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
    echo -e "${RED}Error: Unsupported OS or architecture${NC}"
    echo "Please build from source: https://github.com/${REPO}"
    exit 1
fi

if [ "$OS" = "windows" ]; then
    BINARY_NAME="sqdesk.exe"
fi

# Get latest release
echo -e "${YELLOW}Fetching latest release...${NC}"
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo -e "${YELLOW}No release found. Building from source...${NC}"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        echo "Please install Go from https://golang.org/dl/"
        exit 1
    fi
    
    # Install using go install
    echo -e "${YELLOW}Installing via go install...${NC}"
    go install github.com/${REPO}/cmd/sqdesk@latest
    
    echo -e "${GREEN}✅ SQDesk installed successfully!${NC}"
    echo -e "Run ${BLUE}sqdesk${NC} to start"
    exit 0
fi

echo -e "${YELLOW}Latest version:${NC} ${LATEST_RELEASE}"

# Download URL
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_RELEASE}/sqdesk-${OS}-${ARCH}"

if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
fi

# Create temp directory
TMP_DIR=$(mktemp -d)
TMP_FILE="${TMP_DIR}/${BINARY_NAME}"

echo -e "${YELLOW}Downloading...${NC}"
if curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" 2>/dev/null; then
    chmod +x "$TMP_FILE"
else
    echo -e "${YELLOW}Binary not found. Building from source...${NC}"
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        echo "Please install Go from https://golang.org/dl/"
        rm -rf "$TMP_DIR"
        exit 1
    fi
    
    go install github.com/${REPO}/cmd/sqdesk@latest
    
    echo -e "${GREEN}✅ SQDesk installed successfully!${NC}"
    echo -e "Run ${BLUE}sqdesk${NC} to start"
    rm -rf "$TMP_DIR"
    exit 0
fi

# Install binary
echo -e "${YELLOW}Installing to ${INSTALL_DIR}...${NC}"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo -e "${YELLOW}Requires sudo to install to ${INSTALL_DIR}${NC}"
    sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
fi

# Cleanup
rm -rf "$TMP_DIR"

# Verify installation
if command -v sqdesk &> /dev/null; then
    echo ""
    echo -e "${GREEN}✅ SQDesk installed successfully!${NC}"
    echo ""
    echo -e "Run ${BLUE}sqdesk${NC} to start"
    echo ""
else
    echo -e "${YELLOW}Installation complete. You may need to add ${INSTALL_DIR} to your PATH${NC}"
    echo ""
    echo "Add this to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
fi
