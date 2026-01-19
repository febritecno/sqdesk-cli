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
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        CYGWIN*|MINGW*|MSYS*) echo "windows";;
        *)          echo "unknown";;
    esac
}

# Detect Architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64";;
        arm64|aarch64)  echo "arm64";;
        armv7l)         echo "arm";;
        i386|i686)      echo "386";;
        *)              echo "unknown";;
    esac
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

# Try go install first (most reliable)
install_via_go() {
    if command -v go &> /dev/null; then
        echo -e "${YELLOW}Installing via go install...${NC}"
        go install github.com/${REPO}/cmd/sqdesk@latest
        
        # Check if installed to GOPATH/bin
        GOBIN=$(go env GOPATH)/bin
        if [ -f "${GOBIN}/sqdesk" ]; then
            echo ""
            echo -e "${GREEN}✅ SQDesk installed successfully!${NC}"
            echo ""
            echo -e "Binary location: ${BLUE}${GOBIN}/sqdesk${NC}"
            echo -e "Run ${BLUE}sqdesk${NC} to start"
            
            # Add GOBIN to PATH hint if needed
            if ! command -v sqdesk &> /dev/null; then
                echo ""
                echo -e "${YELLOW}Note: Add this to your shell profile:${NC}"
                echo "  export PATH=\"\$PATH:${GOBIN}\""
            fi
            return 0
        fi
    fi
    return 1
}

# Try downloading pre-built binary
install_via_binary() {
    echo -e "${YELLOW}Fetching latest release...${NC}"
    
    # Get latest release tag
    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$LATEST_RELEASE" ]; then
        echo -e "${YELLOW}No release found.${NC}"
        return 1
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
    
    echo -e "${YELLOW}Downloading from: ${DOWNLOAD_URL}${NC}"
    
    if curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" 2>/dev/null; then
        chmod +x "$TMP_FILE"
        
        # Install binary
        echo -e "${YELLOW}Installing to ${INSTALL_DIR}...${NC}"
        
        if [ -w "$INSTALL_DIR" ]; then
            mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        else
            echo -e "${YELLOW}Requires sudo to install to ${INSTALL_DIR}${NC}"
            sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        fi
        
        rm -rf "$TMP_DIR"
        
        echo ""
        echo -e "${GREEN}✅ SQDesk installed successfully!${NC}"
        echo ""
        echo -e "Run ${BLUE}sqdesk${NC} to start"
        return 0
    else
        rm -rf "$TMP_DIR"
        echo -e "${YELLOW}Binary not found for ${OS}/${ARCH}${NC}"
        return 1
    fi
}

# Main installation logic
echo ""

# Try binary download first, fallback to go install
if install_via_binary; then
    exit 0
fi

echo ""
echo -e "${YELLOW}Falling back to go install...${NC}"

if install_via_go; then
    exit 0
fi

# Neither method worked
echo ""
echo -e "${RED}Installation failed.${NC}"
echo ""
echo "Please install manually:"
echo "  1. Install Go: https://golang.org/dl/"
echo "  2. Run: go install github.com/${REPO}/cmd/sqdesk@latest"
echo ""
echo "Or build from source:"
echo "  git clone https://github.com/${REPO}.git"
echo "  cd sqdesk-cli"
echo "  go build -o sqdesk ./cmd/sqdesk"
exit 1
