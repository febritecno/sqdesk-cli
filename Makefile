.PHONY: build run clean test build-all

# Binary name
BINARY_NAME=sqdesk

# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main package
MAIN_PKG=./cmd/sqdesk

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PKG)

# Run the application
run:
	$(GOCMD) run $(MAIN_PKG)

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Update dependencies
deps:
	$(GOMOD) tidy

# Build for all platforms
build-all: clean
	mkdir -p $(BUILD_DIR)
	# macOS (Intel)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)
	# macOS (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)
	# Linux (amd64)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)
	# Linux (arm64)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PKG)
	# Windows (amd64)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PKG)

# Install the binary to GOPATH/bin
install:
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PKG)

# Development build with race detector
dev:
	$(GOBUILD) -race -o $(BINARY_NAME) $(MAIN_PKG)

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  run        - Run the application"
	@echo "  clean      - Clean build files"
	@echo "  test       - Run tests"
	@echo "  deps       - Update dependencies"
	@echo "  build-all  - Build for all platforms"
	@echo "  install    - Install to GOPATH/bin"
	@echo "  dev        - Development build with race detector"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
