# Bifrost - Package Manager for Carrion

Bifrost is the official package manager for the Carrion programming language. It manages dependencies, handles package installation, and integrates seamlessly with Carrion's import system.

## Installation

### Quick Install (Recommended)

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/javanhut/bifrost/main/scripts/install.sh | bash
```

**Windows (PowerShell as Administrator):**
```powershell
irm https://raw.githubusercontent.com/javanhut/bifrost/main/scripts/install.ps1 | iex
```

### Using Make (Development)

```bash
git clone https://github.com/javanhut/bifrost.git
cd bifrost
make install
```

### Manual Installation

**Prerequisites:**
- Go 1.21 or later
- Git

**Linux/macOS:**
```bash
git clone https://github.com/javanhut/bifrost.git
cd bifrost
go build -o bifrost ./cmd/bifrost
sudo mv bifrost /usr/local/bin/
```

**Windows:**
```cmd
git clone https://github.com/javanhut/bifrost.git
cd bifrost
go build -o bifrost.exe ./cmd/bifrost
# Move bifrost.exe to a directory in your PATH
```

### Platform-Specific Packages

Download pre-built binaries from the [releases page](https://github.com/javanhut/bifrost/releases):

- Linux: `bifrost-VERSION-linux-amd64.tar.gz`
- macOS: `bifrost-VERSION-darwin-amd64.tar.gz` (Intel) or `bifrost-VERSION-darwin-arm64.tar.gz` (Apple Silicon)
- Windows: `bifrost-VERSION-windows-amd64.zip`

### Makefile Commands

```bash
make help           # Show all available commands
make build          # Build for current platform
make install        # Build and install to system
make build-all      # Build for all platforms
make release        # Create release archives
make clean          # Clean build artifacts
make uninstall      # Remove installed binary
```

## Quick Start

### Initialize a New Package

```bash
bifrost init
```

This creates:
- `Bifrost.toml` - Package manifest
- `src/main.crl` - Main module file
- Basic directory structure

### Package Manifest

Example `Bifrost.toml`:

```toml
[package]
name = "my-awesome-lib"
version = "0.1.0"
authors = ["Your Name <you@example.com>"]
description = "An awesome Carrion library"
license = "MIT"
repository = "https://github.com/username/my-awesome-lib"
keywords = ["awesome", "library"]

[package.metadata]
main = "src/main.crl"
include = ["src/**/*.crl", "README.md", "LICENSE"]
exclude = ["tests/**/*", "*.log"]

[dependencies]
json-utils = "^0.3.5"
http-client = ">=1.0.0, <2.0.0"

[dev-dependencies]
test-framework = "^0.4.0"
```

### Version Constraints

- `^1.2.3` - Compatible with 1.x.x (>=1.2.3, <2.0.0)
- `~1.2.3` - Approximately equivalent (>=1.2.3, <1.3.0)
- `1.2.3` - Exact version
- `>=1.2.3` - Minimum version
- `>=1.2.3, <2.0.0` - Version range

## Commands

### `bifrost init`
Create a new Carrion package with a default structure.

### `bifrost install`
Install all dependencies from `Bifrost.toml`.

### `bifrost install <package>`
Install a specific package (registry support coming soon).

### `bifrost list`
List all installed packages.

### `bifrost info`
Display information about the current package.

### `bifrost search <query>`
Search for packages in the registry (coming soon).

### `bifrost version`
Show Bifrost version.

## Package Structure

Standard Carrion package layout:

```
my-package/
├── Bifrost.toml          # Package manifest
├── src/                  # Source files
│   ├── main.crl         # Main module
│   └── lib/             # Library modules
├── tests/               # Test files
├── docs/                # Documentation
└── README.md           # Package documentation
```

## Import Resolution

Bifrost extends Carrion's import system with the following search order:

1. Current directory
2. Project's `carrion_modules/` directory
3. Global packages directory (`~/.carrion/packages/`)
4. Standard library (embedded in Carrion binary)

### Using Installed Packages

```carrion
# Import from an installed package
import "json-utils/parser"

# Import with alias
import "http-client/request" as http
```

## Configuration

### Environment Variables

- `CARRION_HOME` - Override default Carrion home directory (default: `~/.carrion`)
- `CARRION_IMPORT_PATH` - Additional import paths (colon-separated)

### Directory Structure

```
~/.carrion/
├── packages/            # Installed packages
│   ├── json-utils/
│   │   ├── 0.3.5/      # Version directories
│   │   └── 0.3.6/
│   └── http-client/
│       └── 1.2.0/
├── cache/              # Downloaded package archives
└── registry/           # Registry metadata cache
```

## Development Status

Bifrost is in active development. Current features:

- ✅ Package manifest specification
- ✅ Basic CLI commands
- ✅ Version constraint parsing
- ✅ Dependency resolution algorithm
- ✅ Local package management
- ✅ Integration with Carrion imports
- ⏳ Package registry (coming soon)
- ⏳ Package publishing (coming soon)
- ⏳ Private registries (planned)
- ⏳ Workspace support (planned)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

Bifrost is licensed under the MIT License. See LICENSE for details.