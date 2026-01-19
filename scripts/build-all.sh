#!/bin/bash

# Build script for all platforms
# Run this before creating a GitHub release

set -e

VERSION=${1:-"dev"}
OUTPUT_DIR="dist"
BINARY_NAME="sqdesk"

echo "🔨 Building SQDesk ${VERSION}..."
echo ""

# Clean output directory
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# Platforms to build
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "linux/arm"
    "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    
    OUTPUT_NAME="${BINARY_NAME}-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    echo "📦 Building ${OUTPUT_NAME}..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w -X main.Version=${VERSION}" -o "${OUTPUT_DIR}/${OUTPUT_NAME}" ./cmd/sqdesk
    
    if [ $? -eq 0 ]; then
        echo "   ✅ ${OUTPUT_NAME}"
    else
        echo "   ❌ Failed to build ${OUTPUT_NAME}"
    fi
done

echo ""
echo "🎉 Build complete! Binaries are in ${OUTPUT_DIR}/"
echo ""
ls -la $OUTPUT_DIR/

echo ""
echo "📝 Next steps:"
echo "   1. Create a new release on GitHub"
echo "   2. Upload all files from ${OUTPUT_DIR}/"
echo "   3. Tag format: ${VERSION}"
