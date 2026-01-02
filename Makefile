# WTE - Makefile
# ============================================================================

# Variables
BINARY_NAME=wte
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-X wte/internal/cli.Version=$(VERSION) -X wte/internal/cli.BuildTime=$(BUILD_TIME) -X wte/internal/cli.GitCommit=$(GIT_COMMIT) -s -w"

# Directories
CMD_DIR=./cmd/wte
BUILD_DIR=./build
DIST_DIR=./dist

# Platforms for cross-compilation
PLATFORMS=linux/amd64 linux/arm64 linux/arm

.PHONY: all build clean test lint fmt vet deps tidy help install uninstall cross-compile

# Default target
all: clean deps build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for development (with debug info)
build-dev:
	@echo "Building $(BINARY_NAME) (development)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Cross-compile for all platforms
cross-compile:
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} $(CMD_DIR); \
		echo "Built: $(DIST_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}"; \
	done

# Build for Linux amd64
build-linux-amd64:
	@echo "Building for Linux amd64..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@echo "Built: $(DIST_DIR)/$(BINARY_NAME)-linux-amd64"

# Build for Linux arm64
build-linux-arm64:
	@echo "Building for Linux arm64..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	@echo "Built: $(DIST_DIR)/$(BINARY_NAME)-linux-arm64"

# Install to system
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installed: /usr/local/bin/$(BINARY_NAME)"

# Uninstall from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstalled"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned"

# Run the application
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Run with arguments
run-args: build
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# Development: watch for changes and rebuild
watch:
	@echo "Watching for changes..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Run: go install github.com/cosmtrek/air@latest"; \
	fi

# Create release archives
release: cross-compile
	@echo "Creating release archives..."
	@mkdir -p $(DIST_DIR)/release
	@for platform in $(PLATFORMS); do \
		tar -czf $(DIST_DIR)/release/$(BINARY_NAME)-$(VERSION)-$${platform%/*}-$${platform#*/}.tar.gz \
			-C $(DIST_DIR) $(BINARY_NAME)-$${platform%/*}-$${platform#*/}; \
		echo "Created: $(DIST_DIR)/release/$(BINARY_NAME)-$(VERSION)-$${platform%/*}-$${platform#*/}.tar.gz"; \
	done

# Help
help:
	@echo "WTE - Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build the application"
	@echo "  make build        Build the binary"
	@echo "  make build-dev    Build with debug info"
	@echo "  make install      Install to /usr/local/bin"
	@echo "  make uninstall    Remove from /usr/local/bin"
	@echo "  make test         Run tests"
	@echo "  make test-coverage Run tests with coverage"
	@echo "  make fmt          Format code"
	@echo "  make vet          Run go vet"
	@echo "  make lint         Run linter"
	@echo "  make deps         Download dependencies"
	@echo "  make tidy         Tidy dependencies"
	@echo "  make clean        Clean build artifacts"
	@echo "  make cross-compile Build for all platforms"
	@echo "  make release      Create release archives"
	@echo "  make version      Show version info"
	@echo "  make help         Show this help"
