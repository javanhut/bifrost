package config

import (
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	HomeDir      string
	PackagesDir  string
	CacheDir     string
	RegistryDir  string
	ModulesDir   string
}

func New() (*Config, error) {
	homeDir, err := getCarrionHome()
	if err != nil {
		return nil, err
	}

	return &Config{
		HomeDir:      homeDir,
		PackagesDir:  filepath.Join(homeDir, "packages"),
		CacheDir:     filepath.Join(homeDir, "cache"),
		RegistryDir:  filepath.Join(homeDir, "registry"),
		ModulesDir:   "carrion_modules",
	}, nil
}

func getCarrionHome() (string, error) {
	// Check CARRION_HOME environment variable
	if home := os.Getenv("CARRION_HOME"); home != "" {
		return home, nil
	}

	// Use user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userHome, ".carrion"), nil
}

func (c *Config) Init() error {
	// Create all necessary directories
	dirs := []string{
		c.HomeDir,
		c.PackagesDir,
		c.CacheDir,
		c.RegistryDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) PackagePath(name, version string) string {
	return filepath.Join(c.PackagesDir, name, version)
}

func (c *Config) CachePath(filename string) string {
	return filepath.Join(c.CacheDir, filename)
}

func (c *Config) LocalModulesPath() string {
	return c.ModulesDir
}

// GetImportPaths returns the directories to search for imports
func (c *Config) GetImportPaths(workingDir string) []string {
	paths := []string{
		// Current working directory
		workingDir,
		// Local project modules
		filepath.Join(workingDir, c.ModulesDir),
		// Global packages
		c.PackagesDir,
	}

	// Add platform-specific paths
	if runtime.GOOS == "windows" {
		if programData := os.Getenv("ProgramData"); programData != "" {
			paths = append(paths, filepath.Join(programData, "Carrion", "packages"))
		}
	} else {
		// Unix-like systems
		paths = append(paths, "/usr/local/share/carrion/packages")
	}

	return paths
}