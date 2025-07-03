# Bifrost Installation Guide

This guide provides multiple installation methods for Bifrost, the package manager for the Carrion programming language.

## Quick Installation

### One-Line Install (Recommended)

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/javanhut/bifrost/main/install.sh | bash
```

**Windows (PowerShell as Administrator):**
```powershell
irm https://raw.githubusercontent.com/javanhut/bifrost/main/install.ps1 | iex
```

**Windows (User Installation):**
```powershell
irm https://raw.githubusercontent.com/javanhut/bifrost/main/install.ps1 | iex -Global:$false
```

## Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/javanhut/bifrost/releases):

### Linux
```bash
# x86_64
wget https://github.com/javanhut/bifrost/releases/download/v0.1.0/bifrost-0.1.0-linux-amd64.tar.gz
tar -xzf bifrost-0.1.0-linux-amd64.tar.gz
sudo install -m 755 bifrost /usr/local/bin/

# ARM64
wget https://github.com/javanhut/bifrost/releases/download/v0.1.0/bifrost-0.1.0-linux-arm64.tar.gz
tar -xzf bifrost-0.1.0-linux-arm64.tar.gz
sudo install -m 755 bifrost /usr/local/bin/
```

### macOS
```bash
# Intel Mac
wget https://github.com/javanhut/bifrost/releases/download/v0.1.0/bifrost-0.1.0-darwin-amd64.tar.gz
tar -xzf bifrost-0.1.0-darwin-amd64.tar.gz
sudo install -m 755 bifrost /usr/local/bin/

# Apple Silicon (M1/M2)
wget https://github.com/javanhut/bifrost/releases/download/v0.1.0/bifrost-0.1.0-darwin-arm64.tar.gz
tar -xzf bifrost-0.1.0-darwin-arm64.tar.gz
sudo install -m 755 bifrost /usr/local/bin/
```

### Windows
```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/javanhut/bifrost/releases/download/v0.1.0/bifrost-0.1.0-windows-amd64.zip" -OutFile "bifrost.zip"
Expand-Archive -Path "bifrost.zip" -DestinationPath "."

# Move to a directory in your PATH (as Administrator)
Move-Item "bifrost.exe" "C:\Program Files\Bifrost\bifrost.exe"
```

## Build from Source

### Prerequisites
- Go 1.21 or later
- Git
- Make (optional, for easier building)

### Using Make (Recommended)
```bash
git clone https://github.com/javanhut/bifrost.git
cd bifrost
make install
```

Available Make commands:
```bash
make help           # Show all available commands
make build          # Build for current platform
make install        # Build and install to system
make build-all      # Build for all platforms
make release        # Create release archives
make clean          # Clean build artifacts
make test           # Run tests
make fmt            # Format code
make lint           # Run linter
make uninstall      # Remove installed binary
```

### Manual Build
```bash
git clone https://github.com/javanhut/bifrost.git
cd bifrost
go build -o bifrost ./cmd/bifrost

# Linux/macOS
sudo mv bifrost /usr/local/bin/

# Windows
# Move bifrost.exe to a directory in your PATH
```

## Installation Scripts

### Linux/macOS Script
The `install.sh` script supports several environment variables:

```bash
# Custom version
BIFROST_VERSION=0.2.0 ./install.sh

# Custom install directory
INSTALL_DIR=/opt/bifrost/bin ./install.sh

# Custom Carrion home
BIFROST_HOME=/opt/carrion ./install.sh
```

### Windows PowerShell Script
The `install.ps1` script supports parameters:

```powershell
# Install specific version
.\install.ps1 -Version "0.2.0"

# Install to custom directory
.\install.ps1 -InstallDir "C:\Tools\Bifrost"

# Global installation (requires admin)
.\install.ps1 -Global

# Force reinstallation
.\install.ps1 -Force
```

## Package Managers

### Homebrew (macOS/Linux) - Coming Soon
```bash
brew install bifrost
```

### Chocolatey (Windows) - Coming Soon
```powershell
choco install bifrost
```

### Scoop (Windows) - Coming Soon
```powershell
scoop install bifrost
```

## Docker

### Using Docker
```bash
# Build Docker image
docker build -t bifrost:latest .

# Run Bifrost in container
docker run --rm -v $(pwd):/workspace -w /workspace bifrost:latest init
```

### Docker Compose
```yaml
version: '3.8'
services:
  bifrost:
    build: .
    volumes:
      - .:/workspace
    working_dir: /workspace
```

## Verification

After installation, verify Bifrost is working:

```bash
# Check version
bifrost version

# Check help
bifrost --help

# Initialize a test package
mkdir test-package && cd test-package
bifrost init
```

Expected output:
```
Bifrost 0.1.0
Package manager for the Carrion programming language
```

## Environment Setup

### Required Environment Variables
```bash
# Optional: Custom Carrion home directory
export CARRION_HOME="$HOME/.carrion"

# Optional: Additional import paths
export CARRION_IMPORT_PATH="/usr/local/share/carrion/packages:$HOME/.carrion/packages"
```

### Shell Completion (Coming Soon)
```bash
# Bash
bifrost completion bash > /etc/bash_completion.d/bifrost

# Zsh
bifrost completion zsh > /usr/local/share/zsh/site-functions/_bifrost

# Fish
bifrost completion fish > ~/.config/fish/completions/bifrost.fish
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   # Linux/macOS: Use sudo for system installation
   sudo make install
   
   # Windows: Run PowerShell as Administrator
   ```

2. **Command Not Found**
   ```bash
   # Ensure install directory is in PATH
   echo $PATH
   export PATH="$PATH:/usr/local/bin"
   ```

3. **Go Not Found**
   ```bash
   # Install Go from https://golang.org/dl/
   # Or use your package manager:
   sudo apt install golang-go    # Ubuntu/Debian
   brew install go               # macOS
   choco install golang          # Windows
   ```

4. **Build Failures**
   ```bash
   # Clean and rebuild
   make clean
   go mod tidy
   make build
   ```

### Getting Help

- GitHub Issues: https://github.com/javanhut/bifrost/issues
- Documentation: https://github.com/javanhut/bifrost/blob/main/README.md
- Discord: [Carrion Language Community]

## Uninstallation

### Using Make
```bash
make uninstall
```

### Manual Removal
```bash
# Remove binary
sudo rm /usr/local/bin/bifrost

# Remove Carrion directory (optional)
rm -rf ~/.carrion

# Windows
del "C:\Program Files\Bifrost\bifrost.exe"
rmdir "C:\Users\%USERNAME%\.carrion" /s
```