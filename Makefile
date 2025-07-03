# Bifrost Makefile
# Package manager for the Carrion programming language

# Variables
BINARY_NAME := bifrost
VERSION := 0.1.0
BUILD_DIR := build
INSTALL_DIR := /usr/local/bin
GO := go
GOFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

# Detect OS
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# OS-specific settings
ifeq ($(UNAME_S),Linux)
OS := linux
INSTALL_CMD := sudo install -m 755
endif
ifeq ($(UNAME_S),Darwin)
OS := darwin
INSTALL_CMD := sudo install -m 755
endif
ifeq ($(OS),Windows_NT)
OS := windows
BINARY_NAME := bifrost.exe
INSTALL_DIR := C:/Program Files/Bifrost
INSTALL_CMD := copy
endif

# Architecture mapping
ifeq ($(UNAME_M),x86_64)
ARCH := amd64
endif
ifeq ($(UNAME_M),aarch64)
ARCH := arm64
endif
ifeq ($(UNAME_M),arm64)
ARCH := arm64
endif
ifeq ($(UNAME_M),i386)
ARCH := 386
endif
ifeq ($(UNAME_M),i686)
ARCH := 386
endif

# Default target
.DEFAULT_GOAL := build

# Help target
.PHONY: help
help:
	@echo "Bifrost Package Manager - Build and Installation"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build for current platform"
	@echo "  make install      Build and install to system"
	@echo "  make build        Build for current platform"
	@echo "  make build-all    Build for all platforms"
	@echo "  make clean        Remove build artifacts"
	@echo "  make test         Run tests"
	@echo "  make fmt          Format code"
	@echo "  make lint         Run linter"
	@echo "  make uninstall    Remove installed binary"
	@echo ""
	@echo "Platform-specific:"
	@echo "  make build-linux    Build for Linux (amd64)"
	@echo "  make build-darwin   Build for macOS (amd64)"
	@echo "  make build-windows  Build for Windows (amd64)"
	@echo "  make build-arm      Build for ARM platforms"
	@echo ""
	@echo "Current platform: $(OS)/$(ARCH)"

# Build for current platform
.PHONY: build
build:
	@echo "Building Bifrost $(VERSION) for $(OS)/$(ARCH)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/bifrost
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install to system
.PHONY: install
install: build
	@echo "Installing Bifrost to $(INSTALL_DIR)..."
ifeq ($(OS),windows)
	@if not exist "$(INSTALL_DIR)" mkdir "$(INSTALL_DIR)"
	@$(INSTALL_CMD) $(BUILD_DIR)\$(BINARY_NAME) "$(INSTALL_DIR)\$(BINARY_NAME)"
	@echo "Bifrost installed. Add $(INSTALL_DIR) to your PATH."
else
	@$(INSTALL_CMD) $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Bifrost installed to $(INSTALL_DIR)/$(BINARY_NAME)"
endif

# Quick install script for Unix-like systems
.PHONY: quick-install
quick-install:
	@echo "Quick installing Bifrost..."
	@$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/bifrost
	@chmod +x $(BINARY_NAME)
	@echo "Binary built. Run 'sudo make install' to install system-wide"

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows build-arm

# Platform-specific builds
.PHONY: build-linux
build-linux:
	@echo "Building for Linux amd64..."
	@mkdir -p $(BUILD_DIR)/linux-amd64
	@GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/linux-amd64/bifrost ./cmd/bifrost

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)/darwin-amd64
	@GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/darwin-amd64/bifrost ./cmd/bifrost
	@mkdir -p $(BUILD_DIR)/darwin-arm64
	@GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/darwin-arm64/bifrost ./cmd/bifrost

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)/windows-amd64
	@GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/windows-amd64/bifrost.exe ./cmd/bifrost

.PHONY: build-arm
build-arm:
	@echo "Building for ARM platforms..."
	@mkdir -p $(BUILD_DIR)/linux-arm64
	@GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/linux-arm64/bifrost ./cmd/bifrost
	@mkdir -p $(BUILD_DIR)/linux-arm
	@GOOS=linux GOARCH=arm $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/linux-arm/bifrost ./cmd/bifrost

# Development targets
.PHONY: run
run:
	@$(GO) run ./cmd/bifrost

.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test -v ./...

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"

.PHONY: vet
vet:
	@echo "Running go vet..."
	@$(GO) vet ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Uninstall from system
.PHONY: uninstall
uninstall:
	@echo "Uninstalling Bifrost..."
ifeq ($(OS),windows)
	@del "$(INSTALL_DIR)\$(BINARY_NAME)" 2>nul || echo "Bifrost not found in $(INSTALL_DIR)"
else
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Bifrost uninstalled from $(INSTALL_DIR)"
endif

# Dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

# Release targets
.PHONY: release
release: clean build-all
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/releases
	@cd $(BUILD_DIR)/linux-amd64 && tar -czf ../releases/bifrost-$(VERSION)-linux-amd64.tar.gz bifrost
	@cd $(BUILD_DIR)/linux-arm64 && tar -czf ../releases/bifrost-$(VERSION)-linux-arm64.tar.gz bifrost
	@cd $(BUILD_DIR)/darwin-amd64 && tar -czf ../releases/bifrost-$(VERSION)-darwin-amd64.tar.gz bifrost
	@cd $(BUILD_DIR)/darwin-arm64 && tar -czf ../releases/bifrost-$(VERSION)-darwin-arm64.tar.gz bifrost
	@cd $(BUILD_DIR)/windows-amd64 && zip ../releases/bifrost-$(VERSION)-windows-amd64.zip bifrost.exe
	@echo "Release archives created in $(BUILD_DIR)/releases/"

# Version check
.PHONY: version
version:
	@echo "Current version: $(VERSION)"
	@echo "To update version, edit VERSION variable in Makefile"