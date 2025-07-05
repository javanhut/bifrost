# Bifrost - Package Manager for Carrion

Bifrost is the official package manager for the Carrion programming language. It manages dependencies, handles package installation, publishing, and integrates seamlessly with Carrion's import system.

## Features

- 📦 **Package Management** - Install, update, and remove packages effortlessly
- 🌍 **Registry Integration** - Publish and install packages from the Carrion registry
- ⚙️ **Easy Configuration** - Built-in config system for credentials and settings
- 🔄 **Version Control** - Semantic versioning with flexible constraint resolution
- 🏗️ **Multi-tier Installation** - Local, user, and global package scopes
- 🔍 **Package Discovery** - Search and explore available packages
- 🔐 **Secure Publishing** - Authentication support for package publishing

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

### From Source

**Prerequisites:**
- Go 1.21 or later
- Git

```bash
git clone https://github.com/javanhut/bifrost.git
cd bifrost
go build -o bifrost ./cmd/bifrost
sudo mv bifrost /usr/local/bin/  # Linux/macOS
# Or move to a directory in your PATH on Windows
```

### Verify Installation

```bash
bifrost version
```

## Quick Start

### 1. Initialize a New Package

```bash
mkdir my-package && cd my-package
bifrost init
```

This creates:
- `Bifrost.toml` - Package manifest
- `src/main.crl` - Main module file
- Basic directory structure (`src/`, `tests/`, `docs/`)

### 2. Configure Authentication (For Publishing)

```bash
# Set registry credentials
bifrost config set registry.username your-username
bifrost config set registry.password your-password

# Or use API key authentication
bifrost config set registry.api-key your-api-key

# Set user information
bifrost config set user.name "Your Name"
bifrost config set user.email "you@example.com"
```

### 3. Install Dependencies

```bash
# Install packages
bifrost install json-utils
bifrost install http-client@1.2.0

# Install from manifest
bifrost install
```

### 4. Publish Your Package

```bash
bifrost publish
```

## Commands Reference

### Package Creation

#### `bifrost init`
Create a new Carrion package with default structure.

```bash
bifrost init
```

Creates:
- `Bifrost.toml` with package metadata
- `src/main.crl` with sample code
- Standard directory structure

### Package Installation

#### `bifrost install`
Install all dependencies from `Bifrost.toml`.

```bash
bifrost install
```

#### `bifrost install <package>[@version]`
Install a specific package from the registry.

```bash
bifrost install json-utils         # Latest version
bifrost install json-utils@1.2.3   # Specific version
bifrost install json-utils@^1.0.0  # Version constraint
```

#### Global Installation
Install packages system-wide for all users.

```bash
bifrost install --global json-utils
bifrost install -g json-utils        # Short form
```

**Installation Scopes:**
- **Local**: `./carrion_modules/` (project-specific)
- **User**: `~/.carrion/packages/` (user-specific)  
- **Global**: `/usr/local/share/carrion/lib/` (system-wide)

### Package Removal

#### `bifrost uninstall [package][@version]`
Remove packages or all dependencies.

```bash
bifrost uninstall                    # Remove all dependencies
bifrost uninstall json-utils         # Remove specific package
bifrost uninstall json-utils@1.2.3   # Remove specific version
bifrost uninstall --all json-utils   # Remove all versions
bifrost uninstall --global json-utils # Remove global package
```

#### Cache Management
```bash
bifrost uninstall --clean           # Clean package cache
```

### Package Discovery

#### `bifrost search <query>`
Search for packages in the registry.

```bash
bifrost search json
bifrost search http-client
```

#### `bifrost info [package][@version]`
Display package information.

```bash
bifrost info                        # Current package info
bifrost info json-utils            # Latest version from registry
bifrost info json-utils@1.2.3      # Specific version info
```

#### `bifrost list [--global]`
List installed packages.

```bash
bifrost list                        # User packages
bifrost list --global              # Global packages
```

### Package Publishing

#### `bifrost publish`
Publish the current package to the registry.

```bash
bifrost publish
```

**Requirements:**
- Complete `Bifrost.toml` manifest
- Configured authentication credentials
- Package archive will be created automatically

### Configuration Management

#### `bifrost config set <key> <value>`
Set configuration values.

```bash
# Registry settings
bifrost config set registry.url https://registry.carrionlang.com
bifrost config set registry.username your-username
bifrost config set registry.password your-password
bifrost config set registry.api-key your-api-key
bifrost config set registry.auth-type basic  # or 'token', 'none'

# User information
bifrost config set user.name "Your Name"
bifrost config set user.email "you@example.com"
```

#### `bifrost config get [key]`
View configuration values.

```bash
bifrost config get                  # Show all configuration
bifrost config get registry.username # Show specific value
```

#### `bifrost config unset <key>`
Remove configuration values.

```bash
bifrost config unset registry.api-key
bifrost config unset user.email
```

### Authentication (Legacy)

#### `bifrost login`
Interactive authentication (legacy method).

```bash
bifrost login
```

#### `bifrost logout`
Remove stored authentication.

```bash
bifrost logout
```

### Utility Commands

#### `bifrost version`
Show Bifrost version and registry information.

```bash
bifrost version
```

## Package Manifest (Bifrost.toml)

### Basic Structure

```toml
[package]
name = "my-awesome-lib"
version = "0.1.0"
authors = ["Your Name <you@example.com>"]
description = "An awesome Carrion library"
license = "MIT"
repository = "https://github.com/username/my-awesome-lib"
keywords = ["awesome", "library", "carrion"]

[package.metadata]
main = "src/main.crl"
include = ["src/**/*.crl", "README.md", "LICENSE"]
exclude = ["tests/**/*", "*.log", "*.tmp"]

[dependencies]
json-utils = "^0.3.5"
http-client = ">=1.0.0, <2.0.0"
math-helpers = "~1.2.0"

[dev-dependencies]
test-framework = "^0.4.0"
benchmark-utils = "latest"
```

### Version Constraints

| Constraint | Description | Example |
|------------|-------------|---------|
| `^1.2.3` | Compatible with 1.x.x (>=1.2.3, <2.0.0) | `^1.2.3` |
| `~1.2.3` | Approximately equivalent (>=1.2.3, <1.3.0) | `~1.2.3` |
| `1.2.3` | Exact version | `1.2.3` |
| `>=1.2.3` | Minimum version | `>=1.2.3` |
| `>=1.2.3, <2.0.0` | Version range | `>=1.2.3, <2.0.0` |
| `latest` | Latest available version | `latest` |

### Package Fields

#### Required Fields
- `name` - Package name (must be unique in registry)
- `version` - Semantic version (e.g., "1.0.0")
- `authors` - List of authors with optional email
- `description` - Brief package description

#### Optional Fields
- `license` - License identifier (e.g., "MIT", "Apache-2.0")
- `repository` - Source code repository URL
- `homepage` - Package homepage URL
- `keywords` - Array of keywords for discovery

#### Metadata Fields
- `main` - Main module file (default: "src/main.crl")
- `include` - Files to include in package archive
- `exclude` - Files to exclude from package archive

## Configuration

### Configuration System

Bifrost uses a JSON configuration file stored at `~/.carrion/config.json` with the following structure:

```json
{
  "registry": {
    "url": "https://registry.carrionlang.com",
    "username": "your-username",
    "password": "your-password",
    "auth_type": "basic"
  },
  "user": {
    "name": "Your Name",
    "email": "you@example.com"
  }
}
```

### Configuration Precedence

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file**
4. **Default values** (lowest priority)

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CARRION_HOME` | Carrion home directory | `~/.carrion` |
| `CARRION_REGISTRY_URL` | Registry URL | `https://registry.carrionlang.com` |

### Authentication Types

#### Basic Authentication
```bash
bifrost config set registry.auth-type basic
bifrost config set registry.username your-username
bifrost config set registry.password your-password
```

#### Token Authentication
```bash
bifrost config set registry.auth-type token
bifrost config set registry.api-key your-api-key
```

#### No Authentication
```bash
bifrost config set registry.auth-type none
```

## Directory Structure

### Package Layout

Standard Carrion package structure:

```
my-package/
├── Bifrost.toml          # Package manifest
├── src/                  # Source files
│   ├── main.crl         # Main module
│   ├── lib/             # Library modules
│   │   ├── parser.crl
│   │   └── formatter.crl
│   └── utils/           # Utility modules
├── tests/               # Test files
│   ├── main_test.crl
│   └── lib/
├── docs/                # Documentation
│   ├── guide.md
│   └── api.md
├── examples/            # Example code
├── LICENSE              # License file
└── README.md           # Package documentation
```

### Installation Directories

#### User-Specific (Default)
```
~/.carrion/
├── packages/            # Installed packages
│   ├── json-utils/
│   │   ├── 0.3.5/
│   │   └── 0.3.6/
│   └── http-client/
├── cache/              # Package archive cache
├── config.json         # Configuration file
└── auth.json           # Legacy auth file
```

#### Global System-Wide
```
/usr/local/share/carrion/lib/    # Global packages
├── json-utils/
│   └── 1.0.0/
└── http-client/
    └── 2.1.0/
```

#### Project-Local
```
project/
├── Bifrost.toml
├── src/
└── carrion_modules/    # Project dependencies
    ├── test-utils/
    └── dev-helpers/
```

## Import Resolution

Bifrost integrates with Carrion's import system, searching for modules in order:

1. **Current Directory** - Local files
2. **Project Modules** - `./carrion_modules/`
3. **User Packages** - `~/.carrion/packages/`
4. **Global Packages** - `/usr/local/share/carrion/lib/`
5. **Standard Library** - Built-in modules

### Using Packages in Code

```carrion
# Import from installed packages
import "json-utils/parser"
import "http-client/request" as http

# Use imported modules
grimoire JsonProcessor {
    spell process_data(data) {
        parsed = parser.parse(data)
        return parsed
    }
}

main {
    processor = JsonProcessor()
    result = processor.process_data('{"key": "value"}')
    print(result)
}
```

## Registry Integration

### Default Registry

Bifrost connects to the official Carrion Package Registry at `https://registry.carrionlang.com`.

### Publishing Workflow

1. **Prepare Package**
   - Complete `Bifrost.toml` with all required fields
   - Ensure source code is ready
   - Add documentation and examples

2. **Configure Authentication**
   ```bash
   bifrost config set registry.username your-username
   bifrost config set registry.password your-password
   ```

3. **Publish**
   ```bash
   bifrost publish
   ```

4. **Verify Publication**
   ```bash
   bifrost search your-package-name
   ```

### Package Discovery

```bash
# Search by keyword
bifrost search json
bifrost search web
bifrost search testing

# Get package details
bifrost info popular-package
bifrost info specific-package@1.0.0
```

## Security

### Credential Management

- Passwords and API keys are stored securely with restricted file permissions (0600)
- Sensitive values are masked when displayed
- Support for both basic authentication and API key authentication

### Best Practices

1. **Use API Keys** when available for token-based authentication
2. **Set Secure Permissions** on configuration files
3. **Regularly Rotate Credentials** for enhanced security
4. **Use Environment Variables** for CI/CD environments

## Troubleshooting

### Common Issues

#### Authentication Failures
```bash
# Check current configuration
bifrost config get

# Verify credentials
bifrost config get registry.username
bifrost config get registry.auth-type

# Re-configure authentication
bifrost config set registry.password new-password
```

#### Package Not Found
```bash
# Check registry connectivity
bifrost search test

# Verify package name and version
bifrost info package-name
```

#### Installation Failures
```bash
# Clean cache and retry
bifrost uninstall --clean
bifrost install package-name

# Check available disk space and permissions
ls -la ~/.carrion/packages/
```

#### Registry Connection Issues
```bash
# Check registry URL
bifrost config get registry.url

# Test with different registry
bifrost config set registry.url https://alternative-registry.com
```

### Debug Information

```bash
# Show version and configuration
bifrost version

# List all configuration
bifrost config get

# Check installed packages
bifrost list
bifrost list --global
```

## Examples

### Basic Package Development

```bash
# Create new package
mkdir my-json-lib && cd my-json-lib
bifrost init

# Edit Bifrost.toml and source files
# ...

# Test locally
# carrion src/main.crl

# Publish when ready
bifrost config set registry.username myuser
bifrost config set registry.password mypass
bifrost publish
```

### Installing and Using Packages

```bash
# Install dependencies
bifrost install json-utils
bifrost install http-client@^2.0.0

# Use in Carrion code
# import "json-utils/parser"
# import "http-client/request"

# Update dependencies
bifrost install json-utils@latest
```

### Global Package Management

```bash
# Install globally for all projects
bifrost install --global common-utils
bifrost install --global test-framework

# List global packages
bifrost list --global

# Remove global package
bifrost uninstall --global old-package
```

## Development Status

### Current Features ✅

- Complete package management (install, uninstall, list)
- Registry integration (search, info, publish)
- Configuration system with secure credential storage
- Multi-tier installation (local, user, global)
- Version constraint resolution
- Package manifest specification
- Import system integration
- Authentication support (basic auth, API keys)
- Cache management

### Planned Features 🚧

- Private registry support
- Workspace/monorepo support
- Package signing and verification
- Dependency vulnerability scanning
- Package metrics and analytics
- Advanced constraint resolution
- Plugin system

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
git clone https://github.com/javanhut/bifrost.git
cd bifrost
go mod download
go build -o bifrost ./cmd/bifrost
./bifrost version
```

### Running Tests

```bash
go test ./...
```

## License

Bifrost is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Support

- 📖 **Documentation**: [bifrost.carrionlang.com](https://bifrost.carrionlang.com)
- 🐛 **Issues**: [GitHub Issues](https://github.com/javanhut/bifrost/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/javanhut/bifrost/discussions)
- 📧 **Email**: bifrost@carrionlang.com