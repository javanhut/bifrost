#!/bin/bash
# Bifrost Installation Script for Unix-like systems (Linux, macOS)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BIFROST_VERSION="${BIFROST_VERSION:-0.1.0}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BIFROST_HOME="${BIFROST_HOME:-$HOME/.carrion}"

# Functions
print_error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_info() {
    echo -e "${YELLOW}$1${NC}"
}

detect_os() {
    case "$(uname -s)" in
        Linux*)     OS=linux;;
        Darwin*)    OS=darwin;;
        *)          print_error "Unsupported operating system"; exit 1;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64)     ARCH=amd64;;
        aarch64)    ARCH=arm64;;
        arm64)      ARCH=arm64;;
        armv7l)     ARCH=arm;;
        *)          print_error "Unsupported architecture"; exit 1;;
    esac
}

check_dependencies() {
    local deps=("curl" "tar" "sudo")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            print_error "$dep is required but not installed"
            exit 1
        fi
    done
}

check_go() {
    if ! command -v go &> /dev/null; then
        print_info "Go is not installed. Bifrost requires Go to build from source."
        print_info "Visit https://golang.org/dl/ to install Go."
        return 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Found Go version: $go_version"
    return 0
}

download_binary() {
    local url="https://github.com/javanhut/bifrost/releases/download/v${BIFROST_VERSION}/bifrost-${BIFROST_VERSION}-${OS}-${ARCH}.tar.gz"
    local temp_dir=$(mktemp -d)
    
    print_info "Downloading Bifrost ${BIFROST_VERSION} for ${OS}/${ARCH}..."
    
    if curl -L -f -o "${temp_dir}/bifrost.tar.gz" "$url" 2>/dev/null; then
        cd "$temp_dir"
        tar -xzf bifrost.tar.gz
        return 0
    else
        print_info "Binary not available for download. Building from source..."
        return 1
    fi
}

build_from_source() {
    if ! check_go; then
        print_error "Cannot build from source without Go"
        exit 1
    fi
    
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    print_info "Cloning Bifrost repository..."
    git clone https://github.com/javanhut/bifrost.git
    cd bifrost
    
    print_info "Building Bifrost..."
    make build
    
    if [ -f "build/bifrost" ]; then
        cp build/bifrost "$temp_dir/bifrost"
        cd "$temp_dir"
        return 0
    else
        print_error "Build failed"
        exit 1
    fi
}

install_binary() {
    print_info "Installing Bifrost to ${INSTALL_DIR}..."
    
    if [ -w "$INSTALL_DIR" ]; then
        install -m 755 bifrost "$INSTALL_DIR/bifrost"
    else
        sudo install -m 755 bifrost "$INSTALL_DIR/bifrost"
    fi
    
    # Verify installation
    if command -v bifrost &> /dev/null; then
        print_success "Bifrost installed successfully!"
        bifrost version
    else
        print_error "Installation failed. Make sure $INSTALL_DIR is in your PATH"
        exit 1
    fi
}

setup_directories() {
    print_info "Setting up Bifrost directories..."
    
    mkdir -p "$BIFROST_HOME"/{packages,cache,registry}
    
    print_success "Created Bifrost home directory: $BIFROST_HOME"
}

main() {
    echo "Bifrost Package Manager Installer"
    echo "================================"
    echo
    
    # Detect system
    detect_os
    detect_arch
    check_dependencies
    
    print_info "System: ${OS}/${ARCH}"
    print_info "Install directory: ${INSTALL_DIR}"
    
    # Try to download binary first, fall back to building from source
    if download_binary; then
        install_binary
    else
        build_from_source
        install_binary
    fi
    
    # Setup directories
    setup_directories
    
    # Cleanup
    rm -rf "$temp_dir"
    
    echo
    print_success "Installation complete!"
    echo
    echo "To get started:"
    echo "  bifrost init       # Create a new Carrion package"
    echo "  bifrost --help     # Show available commands"
    echo
    echo "Make sure to add Carrion's import paths to your environment:"
    echo "  export CARRION_HOME=$BIFROST_HOME"
}

# Run main function
main "$@"