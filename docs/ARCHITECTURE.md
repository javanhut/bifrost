# Bifrost Architecture

## Overview

Bifrost is the package manager for the Carrion programming language. It manages dependencies, handles package installation, and integrates with Carrion's import system.

## Core Components

### 1. Package Manifest (Bifrost.toml)

Each Carrion package contains a `Bifrost.toml` manifest file that describes the package:

```toml
[package]
name = "example-package"
version = "0.1.0"
authors = ["Your Name <you@example.com>"]
description = "A sample Carrion package"
license = "MIT"
repository = "https://github.com/user/example-package"
keywords = ["example", "carrion"]

[dependencies]
json-utils = "^0.3.5"
http-client = ">=1.0.0, <2.0.0"
math-extensions = "~1.2.3"

[dev-dependencies]
test-framework = "^0.4.0"

[package.metadata]
main = "src/main.crl"
include = ["src/**/*.crl", "README.md", "LICENSE"]
exclude = ["tests/**/*", "*.log"]
```

### 2. Package Structure

Standard Carrion package layout:

```
my-package/
├── Bifrost.toml          # Package manifest
├── src/                  # Source files
│   ├── main.crl         # Main module file
│   └── lib/             # Library modules
├── tests/               # Test files
├── docs/                # Documentation
└── README.md           # Package documentation
```

### 3. Package Installation Directory

Packages are installed in a global directory structure:

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

### 4. Dependency Resolution

Bifrost uses a semantic versioning-aware dependency resolver:

1. **Version Constraints**:
   - `^1.2.3` - Compatible with 1.x.x (>=1.2.3, <2.0.0)
   - `~1.2.3` - Approximately equivalent (>=1.2.3, <1.3.0)
   - `1.2.3` - Exact version
   - `>=1.2.3` - Minimum version
   - `>=1.2.3, <2.0.0` - Version range

2. **Resolution Algorithm**:
   - Build dependency graph
   - Find compatible version sets
   - Prefer newer versions within constraints
   - Detect and report conflicts

### 5. Integration with Carrion

Bifrost extends Carrion's import system:

1. **Import Resolution Order**:
   - Current directory
   - Project's `carrion_modules/` directory
   - Global packages directory (`~/.carrion/packages/`)
   - Standard library (embedded in Carrion binary)

2. **Package Imports**:
   ```carrion
   # Import from installed package
   import "json-utils/parser"
   
   # Import with version (future feature)
   import "http-client@1.2.0/client"
   ```

### 6. CLI Commands

- `bifrost init` - Initialize a new Carrion package
- `bifrost install` - Install dependencies from Bifrost.toml
- `bifrost install <package>` - Install a specific package
- `bifrost update` - Update dependencies to latest compatible versions
- `bifrost remove <package>` - Remove a package
- `bifrost publish` - Publish package to registry
- `bifrost search <query>` - Search for packages
- `bifrost info <package>` - Show package information
- `bifrost list` - List installed packages
- `bifrost clean` - Clean cache

### 7. Package Registry

Future registry will provide:
- Package hosting
- Version management
- Search functionality
- Download statistics
- Security scanning

### 8. Local Development

Support for local package development:
- Link local packages with `bifrost link`
- Override dependencies with local paths
- Workspace support for monorepos

## Implementation Phases

1. **Phase 1** (Current): Basic package management
   - Manifest parsing
   - Local package installation
   - Simple dependency resolution

2. **Phase 2**: Enhanced features
   - Semantic versioning
   - Dependency conflict resolution
   - Local package cache

3. **Phase 3**: Registry integration
   - Package publishing
   - Search functionality
   - Security features

4. **Phase 4**: Advanced features
   - Workspaces
   - Private registries
   - Plugin system