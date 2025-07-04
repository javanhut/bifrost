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

### `bifrost install --global <package>`
Install a package globally to the shared system location (`/usr/local/share/carrion/lib/`).

### `bifrost install -g <package>`
Short form of global installation.

### `bifrost list`
List all installed packages (both local and user-specific).

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

Bifrost integrates with Carrion's import system, which searches for modules in the following order:

1. **Current Directory** - Local files relative to the current working directory
2. **Project Modules** - `./carrion_modules/` directory for project-specific packages
3. **User Packages** - `~/.carrion/packages/` for user-installed packages
4. **Global Packages** - `/usr/local/share/carrion/lib/` for system-wide packages (managed by Bifrost)
5. **Standard Library** - Built-in Munin standard library modules

This multi-tier system allows for flexible package management, from local development to system-wide installations.

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

Bifrost manages packages in multiple locations depending on installation type:

#### User-Specific Packages
```
~/.carrion/
├── packages/            # User-installed packages
│   ├── json-utils/
│   │   ├── 0.3.5/      # Version directories
│   │   └── 0.3.6/
│   └── http-client/
│       └── 1.2.0/
├── cache/              # Downloaded package archives
└── registry/           # Registry metadata cache
```

#### Global System Packages
```
/usr/local/share/carrion/
└── lib/                # Globally installed packages
    ├── json-utils/
    │   ├── 1.0.0/
    │   │   ├── parser.crl
    │   │   ├── formatter.crl
    │   │   └── Bifrost.toml
    │   └── 1.0.1/
    └── http-client/
        └── 2.1.0/
            ├── request.crl
            ├── response.crl
            └── auth.crl
```

#### Project-Specific Packages
```
project/
└── carrion_modules/    # Project-local packages
    ├── test-utils/
    │   ├── mock.crl
    │   └── assert.crl
    └── dev-helpers/
        └── debug.crl
```

## Development Status

Bifrost is in active development. Current features:

- ✅ Package manifest specification
- ✅ Basic CLI commands (`init`, `install`, `list`, `info`, `version`)
- ✅ Global package installation (`--global` flag)
- ✅ Multi-tier import resolution system
- ✅ Version constraint parsing
- ✅ Dependency resolution algorithm
- ✅ Local package management
- ✅ User-specific package management
- ✅ System-wide global package management
- ✅ Full integration with Carrion imports
- ⏳ Package registry (coming soon)
- ⏳ Package publishing (coming soon)
- ⏳ Private registries (planned)
- ⏳ Workspace support (planned)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

Bifrost is licensed under the MIT License. See LICENSE for details.