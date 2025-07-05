package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	HomeDir     string
	PackagesDir string
	CacheDir    string
	RegistryDir string
	ModulesDir  string
	RegistryURL string
	AuthFile    string
	ConfigFile  string
}

type AuthConfig struct {
	APIKey   string `json:"api_key,omitempty"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Registry string `json:"registry"`
	AuthType string `json:"auth_type"` // "basic" or "token"
}

type UserConfig struct {
	Registry   RegistryConfig `json:"registry"`
	User       UserInfo       `json:"user,omitempty"`
}

type RegistryConfig struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	AuthType string `json:"auth_type,omitempty"` // "basic", "token", or "none"
}

type UserInfo struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

func New() (*Config, error) {
	homeDir, err := getCarrionHome()
	if err != nil {
		return nil, err
	}

	registryURL := os.Getenv("CARRION_REGISTRY_URL")
	if registryURL == "" {
		registryURL = "https://registry.carrionlang.com"
	}

	return &Config{
		HomeDir:     homeDir,
		PackagesDir: filepath.Join(homeDir, "packages"),
		CacheDir:    filepath.Join(homeDir, "cache"),
		RegistryDir: filepath.Join(homeDir, "registry"),
		ModulesDir:  "carrion_modules",
		RegistryURL: registryURL,
		AuthFile:    filepath.Join(homeDir, "auth.json"),
		ConfigFile:  filepath.Join(homeDir, "config.json"),
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

// LocalPackagePath returns the path for a local project package
func (c *Config) LocalPackagePath(name, version string) string {
	return filepath.Join(c.ModulesDir, name, version)
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
		// Global packages (user-specific)
		c.PackagesDir,
	}

	// Add platform-specific shared global paths
	if runtime.GOOS == "windows" {
		if programData := os.Getenv("ProgramData"); programData != "" {
			paths = append(paths, filepath.Join(programData, "Carrion", "lib"))
		}
	} else {
		// Unix-like systems - shared global location
		paths = append(paths, "/usr/local/share/carrion/lib")
	}

	return paths
}

// GetSharedGlobalPackagesDir returns the shared global packages directory
func (c *Config) GetSharedGlobalPackagesDir() string {
	if runtime.GOOS == "windows" {
		if programData := os.Getenv("ProgramData"); programData != "" {
			return filepath.Join(programData, "Carrion", "lib")
		}
		return filepath.Join("C:", "ProgramData", "Carrion", "lib")
	}
	return "/usr/local/share/carrion/lib"
}

// LoadAuth loads the authentication configuration from the auth file
func (c *Config) LoadAuth() (*AuthConfig, error) {
	data, err := os.ReadFile(c.AuthFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No auth file exists
		}
		return nil, err
	}

	var auth AuthConfig
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, err
	}

	return &auth, nil
}

// SaveAuth saves the authentication configuration to the auth file
func (c *Config) SaveAuth(auth *AuthConfig) error {
	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.AuthFile, data, 0600) // Secure permissions
}

// ClearAuth removes the authentication configuration file
func (c *Config) ClearAuth() error {
	if err := os.Remove(c.AuthFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// LoadUserConfig loads the user configuration from the config file
func (c *Config) LoadUserConfig() (*UserConfig, error) {
	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &UserConfig{
				Registry: RegistryConfig{
					URL:      c.RegistryURL,
					AuthType: "none",
				},
			}, nil
		}
		return nil, err
	}

	var config UserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Use environment or default registry URL if not set in config
	if config.Registry.URL == "" {
		config.Registry.URL = c.RegistryURL
	}

	return &config, nil
}

// SaveUserConfig saves the user configuration to the config file
func (c *Config) SaveUserConfig(config *UserConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.ConfigFile, data, 0600) // Secure permissions
}

// GetRegistryConfig returns the effective registry configuration (config file + environment)
func (c *Config) GetRegistryConfig() (*RegistryConfig, error) {
	userConfig, err := c.LoadUserConfig()
	if err != nil {
		return nil, err
	}

	registryConfig := userConfig.Registry

	// Override with environment variables if set
	if envURL := os.Getenv("CARRION_REGISTRY_URL"); envURL != "" {
		registryConfig.URL = envURL
	}

	return &registryConfig, nil
}
