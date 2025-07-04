package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/javanhut/bifrost/internal/config"
)

// CarrionIntegration provides methods to integrate with Carrion's import system
type CarrionIntegration struct {
	config *config.Config
}

func New(cfg *config.Config) *CarrionIntegration {
	return &CarrionIntegration{
		config: cfg,
	}
}

// ResolveImport resolves a Carrion import path to a file path
func (ci *CarrionIntegration) ResolveImport(importPath string, workingDir string) (string, error) {
	// Remove .crl extension if present
	importPath = strings.TrimSuffix(importPath, ".crl")

	// Get search paths
	searchPaths := ci.config.GetImportPaths(workingDir)

	for _, basePath := range searchPaths {
		// Try direct path
		fullPath := filepath.Join(basePath, importPath+".crl")
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}

		// For package imports (e.g., "json-utils/parser"), check in package directories
		parts := strings.Split(importPath, "/")
		if len(parts) > 1 && basePath == ci.config.PackagesDir {
			// Look for the latest version of the package
			packageName := parts[0]
			packagePath := filepath.Join(basePath, packageName)

			versions, err := ci.getPackageVersions(packagePath)
			if err == nil && len(versions) > 0 {
				// Use the latest version
				latestVersion := versions[len(versions)-1]
				subPath := strings.Join(parts[1:], "/")
				fullPath = filepath.Join(packagePath, latestVersion, subPath+".crl")

				if _, err := os.Stat(fullPath); err == nil {
					return fullPath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("import not found: %s", importPath)
}

// getPackageVersions returns sorted list of versions for a package
func (ci *CarrionIntegration) getPackageVersions(packagePath string) ([]string, error) {
	entries, err := os.ReadDir(packagePath)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	// TODO: Sort by semantic version
	return versions, nil
}

// GenerateImportConfig creates a configuration file for Carrion to use
func (ci *CarrionIntegration) GenerateImportConfig(projectDir string) error {
	configPath := filepath.Join(projectDir, ".carrion_imports")

	// Get all import paths
	importPaths := ci.config.GetImportPaths(projectDir)

	// Write to file
	content := strings.Join(importPaths, "\n")
	return os.WriteFile(configPath, []byte(content), 0644)
}

// SetupEnvironment sets up environment variables for Carrion
func (ci *CarrionIntegration) SetupEnvironment() error {
	// Set CARRION_IMPORT_PATH environment variable
	importPaths := ci.config.GetImportPaths(".")
	pathString := strings.Join(importPaths, string(os.PathListSeparator))

	return os.Setenv("CARRION_IMPORT_PATH", pathString)
}

// LinkPackage creates a symlink in the local modules directory
func (ci *CarrionIntegration) LinkPackage(packageName, version string) error {
	source := ci.config.PackagePath(packageName, version)
	target := filepath.Join(ci.config.LocalModulesPath(), packageName)

	// Remove existing link
	os.Remove(target)

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	// Create symlink
	return os.Symlink(source, target)
}

// CreateModulesDirectory creates the local modules directory structure
func (ci *CarrionIntegration) CreateModulesDirectory(projectDir string) error {
	modulesDir := filepath.Join(projectDir, ci.config.LocalModulesPath())
	return os.MkdirAll(modulesDir, 0755)
}
