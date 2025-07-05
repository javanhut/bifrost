package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNew(t *testing.T) {
	// Save original environment
	originalHome := os.Getenv("CARRION_HOME")
	originalRegistry := os.Getenv("CARRION_REGISTRY_URL")
	defer func() {
		os.Setenv("CARRION_HOME", originalHome)
		os.Setenv("CARRION_REGISTRY_URL", originalRegistry)
	}()

	tests := []struct {
		name        string
		envHome     string
		envRegistry string
		wantErr     bool
		checkFunc   func(*testing.T, *Config)
	}{
		{
			name:        "default configuration",
			envHome:     "",
			envRegistry: "",
			checkFunc: func(t *testing.T, c *Config) {
				if c.RegistryURL != "https://registry.carrionlang.com" {
					t.Errorf("expected default registry URL, got %s", c.RegistryURL)
				}
				if c.ModulesDir != "carrion_modules" {
					t.Errorf("expected default modules dir, got %s", c.ModulesDir)
				}
			},
		},
		{
			name:        "custom CARRION_HOME",
			envHome:     "/custom/carrion/home",
			envRegistry: "",
			checkFunc: func(t *testing.T, c *Config) {
				if c.HomeDir != "/custom/carrion/home" {
					t.Errorf("expected custom home dir, got %s", c.HomeDir)
				}
				if c.PackagesDir != filepath.Join("/custom/carrion/home", "packages") {
					t.Errorf("expected packages dir under custom home, got %s", c.PackagesDir)
				}
			},
		},
		{
			name:        "custom registry URL",
			envHome:     "",
			envRegistry: "https://custom.registry.com",
			checkFunc: func(t *testing.T, c *Config) {
				if c.RegistryURL != "https://custom.registry.com" {
					t.Errorf("expected custom registry URL, got %s", c.RegistryURL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CARRION_HOME", tt.envHome)
			os.Setenv("CARRION_REGISTRY_URL", tt.envRegistry)

			got, err := New()
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

func TestGetCarrionHome(t *testing.T) {
	// Save original environment
	originalHome := os.Getenv("CARRION_HOME")
	defer os.Setenv("CARRION_HOME", originalHome)

	tests := []struct {
		name    string
		envHome string
		want    string
		wantErr bool
	}{
		{
			name:    "with CARRION_HOME set",
			envHome: "/custom/path",
			want:    "/custom/path",
		},
		{
			name:    "without CARRION_HOME set",
			envHome: "",
			want:    "", // Will be checked differently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CARRION_HOME", tt.envHome)

			got, err := getCarrionHome()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCarrionHome() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.envHome != "" {
				if got != tt.want {
					t.Errorf("getCarrionHome() = %v, want %v", got, tt.want)
				}
			} else {
				// When CARRION_HOME is not set, it should use user home
				userHome, _ := os.UserHomeDir()
				expected := filepath.Join(userHome, ".carrion")
				if got != expected {
					t.Errorf("getCarrionHome() = %v, want %v", got, expected)
				}
			}
		})
	}
}

func TestConfig_Init(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	cfg := &Config{
		HomeDir:     filepath.Join(tempDir, "home"),
		PackagesDir: filepath.Join(tempDir, "packages"),
		CacheDir:    filepath.Join(tempDir, "cache"),
		RegistryDir: filepath.Join(tempDir, "registry"),
	}

	err := cfg.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Check that all directories were created
	dirs := []string{cfg.HomeDir, cfg.PackagesDir, cfg.CacheDir, cfg.RegistryDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory %s was not created", dir)
		}
	}
}

func TestConfig_PackagePath(t *testing.T) {
	cfg := &Config{
		PackagesDir: "/home/user/.carrion/packages",
	}

	tests := []struct {
		name    string
		pkgName string
		version string
		want    string
	}{
		{
			name:    "standard package",
			pkgName: "mypackage",
			version: "1.2.3",
			want:    filepath.Join("/home/user/.carrion/packages", "mypackage", "1.2.3"),
		},
		{
			name:    "scoped package",
			pkgName: "@scope/package",
			version: "0.1.0",
			want:    filepath.Join("/home/user/.carrion/packages", "@scope/package", "0.1.0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.PackagePath(tt.pkgName, tt.version); got != tt.want {
				t.Errorf("PackagePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_LocalPackagePath(t *testing.T) {
	cfg := &Config{
		ModulesDir: "carrion_modules",
	}

	got := cfg.LocalPackagePath("mypackage", "1.2.3")
	want := filepath.Join("carrion_modules", "mypackage", "1.2.3")
	
	if got != want {
		t.Errorf("LocalPackagePath() = %v, want %v", got, want)
	}
}

func TestConfig_CachePath(t *testing.T) {
	cfg := &Config{
		CacheDir: "/home/user/.carrion/cache",
	}

	got := cfg.CachePath("package-1.2.3.tar.gz")
	want := filepath.Join("/home/user/.carrion/cache", "package-1.2.3.tar.gz")
	
	if got != want {
		t.Errorf("CachePath() = %v, want %v", got, want)
	}
}

func TestConfig_LocalModulesPath(t *testing.T) {
	cfg := &Config{
		ModulesDir: "carrion_modules",
	}

	if got := cfg.LocalModulesPath(); got != "carrion_modules" {
		t.Errorf("LocalModulesPath() = %v, want carrion_modules", got)
	}
}

func TestConfig_GetImportPaths(t *testing.T) {
	cfg := &Config{
		ModulesDir:  "carrion_modules",
		PackagesDir: "/home/user/.carrion/packages",
	}

	workingDir := "/project/dir"
	paths := cfg.GetImportPaths(workingDir)

	// Check that all expected paths are present
	expectedPaths := []string{
		workingDir,
		filepath.Join(workingDir, "carrion_modules"),
		"/home/user/.carrion/packages",
	}

	for i, expected := range expectedPaths {
		if i >= len(paths) || paths[i] != expected {
			t.Errorf("expected path[%d] to be %s, got %v", i, expected, paths[i])
		}
	}

	// Check platform-specific path
	if runtime.GOOS == "windows" {
		// Should contain ProgramData path
		found := false
		for _, p := range paths {
			if filepath.Base(filepath.Dir(p)) == "Carrion" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Windows paths should include ProgramData Carrion path")
		}
	} else {
		// Unix-like systems
		if paths[len(paths)-1] != "/usr/local/share/carrion/lib" {
			t.Error("Unix paths should include /usr/local/share/carrion/lib")
		}
	}
}

func TestConfig_GetSharedGlobalPackagesDir(t *testing.T) {
	cfg := &Config{}

	got := cfg.GetSharedGlobalPackagesDir()
	
	if runtime.GOOS == "windows" {
		// On Windows, should return ProgramData path
		if !filepath.IsAbs(got) || !filepath.HasPrefix(filepath.Base(filepath.Dir(got)), "Carrion") {
			t.Errorf("Windows shared dir should be under Carrion, got %s", got)
		}
	} else {
		// Unix-like systems
		want := "/usr/local/share/carrion/lib"
		if got != want {
			t.Errorf("Unix shared dir = %v, want %v", got, want)
		}
	}
}

func TestConfig_LoadAuth(t *testing.T) {
	tempDir := t.TempDir()
	
	tests := []struct {
		name        string
		authFile    string
		authContent string
		want        *AuthConfig
		wantErr     bool
	}{
		{
			name:     "valid auth file",
			authFile: filepath.Join(tempDir, "auth1.json"),
			authContent: `{
				"api_key": "test-key",
				"username": "testuser",
				"registry": "https://registry.test.com",
				"auth_type": "token"
			}`,
			want: &AuthConfig{
				APIKey:   "test-key",
				Username: "testuser",
				Registry: "https://registry.test.com",
				AuthType: "token",
			},
		},
		{
			name:     "non-existent file",
			authFile: filepath.Join(tempDir, "nonexistent.json"),
			want:     nil,
			wantErr:  false,
		},
		{
			name:        "invalid JSON",
			authFile:    filepath.Join(tempDir, "invalid.json"),
			authContent: `{invalid json}`,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{AuthFile: tt.authFile}

			if tt.authContent != "" {
				err := os.WriteFile(tt.authFile, []byte(tt.authContent), 0600)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			got, err := cfg.LoadAuth()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil && tt.want != nil {
				if got.APIKey != tt.want.APIKey || got.Username != tt.want.Username {
					t.Errorf("LoadAuth() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestConfig_SaveAuth(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")
	
	cfg := &Config{AuthFile: authFile}
	
	auth := &AuthConfig{
		APIKey:   "test-key",
		Username: "testuser",
		Password: "testpass",
		Registry: "https://registry.test.com",
		AuthType: "basic",
	}

	err := cfg.SaveAuth(auth)
	if err != nil {
		t.Fatalf("SaveAuth() error = %v", err)
	}

	// Verify file was created with correct permissions
	info, err := os.Stat(authFile)
	if err != nil {
		t.Fatalf("failed to stat auth file: %v", err)
	}
	
	if runtime.GOOS != "windows" {
		// Check permissions on Unix-like systems
		if info.Mode().Perm() != 0600 {
			t.Errorf("auth file permissions = %v, want 0600", info.Mode().Perm())
		}
	}

	// Verify content
	data, err := os.ReadFile(authFile)
	if err != nil {
		t.Fatalf("failed to read auth file: %v", err)
	}

	var saved AuthConfig
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("failed to unmarshal saved auth: %v", err)
	}

	if saved.APIKey != auth.APIKey || saved.Username != auth.Username {
		t.Errorf("saved auth = %v, want %v", saved, auth)
	}
}

func TestConfig_ClearAuth(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")
	
	cfg := &Config{AuthFile: authFile}

	// Create a file to clear
	err := os.WriteFile(authFile, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Clear the auth
	err = cfg.ClearAuth()
	if err != nil {
		t.Fatalf("ClearAuth() error = %v", err)
	}

	// Verify file was removed
	if _, err := os.Stat(authFile); !os.IsNotExist(err) {
		t.Error("auth file should have been removed")
	}

	// Clearing non-existent file should not error
	err = cfg.ClearAuth()
	if err != nil {
		t.Errorf("ClearAuth() on non-existent file error = %v", err)
	}
}

func TestConfig_LoadUserConfig(t *testing.T) {
	tempDir := t.TempDir()
	
	tests := []struct {
		name         string
		configFile   string
		configContent string
		registryURL  string
		want         *UserConfig
		wantErr      bool
	}{
		{
			name:       "valid config file",
			configFile: filepath.Join(tempDir, "config1.json"),
			configContent: `{
				"registry": {
					"url": "https://custom.registry.com",
					"auth_type": "token"
				},
				"user": {
					"name": "Test User",
					"email": "test@example.com"
				}
			}`,
			want: &UserConfig{
				Registry: RegistryConfig{
					URL:      "https://custom.registry.com",
					AuthType: "token",
				},
				User: UserInfo{
					Name:  "Test User",
					Email: "test@example.com",
				},
			},
		},
		{
			name:        "non-existent file returns default",
			configFile:  filepath.Join(tempDir, "nonexistent.json"),
			registryURL: "https://default.registry.com",
			want: &UserConfig{
				Registry: RegistryConfig{
					URL:      "https://default.registry.com",
					AuthType: "none",
				},
			},
		},
		{
			name:          "invalid JSON",
			configFile:    filepath.Join(tempDir, "invalid.json"),
			configContent: `{invalid json}`,
			wantErr:       true,
		},
		{
			name:       "empty registry URL uses default",
			configFile: filepath.Join(tempDir, "config2.json"),
			configContent: `{
				"registry": {
					"auth_type": "basic"
				}
			}`,
			registryURL: "https://default.registry.com",
			want: &UserConfig{
				Registry: RegistryConfig{
					URL:      "https://default.registry.com",
					AuthType: "basic",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ConfigFile:  tt.configFile,
				RegistryURL: tt.registryURL,
			}

			if tt.configContent != "" {
				err := os.WriteFile(tt.configFile, []byte(tt.configContent), 0600)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			got, err := cfg.LoadUserConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadUserConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil && tt.want != nil {
				if got.Registry.URL != tt.want.Registry.URL || got.Registry.AuthType != tt.want.Registry.AuthType {
					t.Errorf("LoadUserConfig() registry = %v, want %v", got.Registry, tt.want.Registry)
				}
				if tt.want.User.Name != "" && got.User.Name != tt.want.User.Name {
					t.Errorf("LoadUserConfig() user = %v, want %v", got.User, tt.want.User)
				}
			}
		})
	}
}

func TestConfig_SaveUserConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	
	cfg := &Config{ConfigFile: configFile}
	
	userConfig := &UserConfig{
		Registry: RegistryConfig{
			URL:      "https://test.registry.com",
			Username: "testuser",
			AuthType: "basic",
		},
		User: UserInfo{
			Name:  "Test User",
			Email: "test@example.com",
		},
	}

	err := cfg.SaveUserConfig(userConfig)
	if err != nil {
		t.Fatalf("SaveUserConfig() error = %v", err)
	}

	// Verify file was created with correct permissions
	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	
	if runtime.GOOS != "windows" {
		// Check permissions on Unix-like systems
		if info.Mode().Perm() != 0600 {
			t.Errorf("config file permissions = %v, want 0600", info.Mode().Perm())
		}
	}

	// Verify content
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var saved UserConfig
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("failed to unmarshal saved config: %v", err)
	}

	if saved.Registry.URL != userConfig.Registry.URL {
		t.Errorf("saved registry URL = %v, want %v", saved.Registry.URL, userConfig.Registry.URL)
	}
	if saved.User.Name != userConfig.User.Name {
		t.Errorf("saved user name = %v, want %v", saved.User.Name, userConfig.User.Name)
	}
}

func TestConfig_GetRegistryConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	
	// Save original environment
	originalURL := os.Getenv("CARRION_REGISTRY_URL")
	defer os.Setenv("CARRION_REGISTRY_URL", originalURL)

	tests := []struct {
		name          string
		configContent string
		envURL        string
		defaultURL    string
		want          *RegistryConfig
	}{
		{
			name: "config file URL",
			configContent: `{
				"registry": {
					"url": "https://config.registry.com",
					"auth_type": "token"
				}
			}`,
			defaultURL: "https://default.registry.com",
			want: &RegistryConfig{
				URL:      "https://config.registry.com",
				AuthType: "token",
			},
		},
		{
			name: "environment overrides config",
			configContent: `{
				"registry": {
					"url": "https://config.registry.com",
					"auth_type": "token"
				}
			}`,
			envURL:     "https://env.registry.com",
			defaultURL: "https://default.registry.com",
			want: &RegistryConfig{
				URL:      "https://env.registry.com",
				AuthType: "token",
			},
		},
		{
			name:       "default URL when no config",
			defaultURL: "https://default.registry.com",
			want: &RegistryConfig{
				URL:      "https://default.registry.com",
				AuthType: "none",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CARRION_REGISTRY_URL", tt.envURL)

			cfg := &Config{
				ConfigFile:  configFile,
				RegistryURL: tt.defaultURL,
			}

			if tt.configContent != "" {
				err := os.WriteFile(configFile, []byte(tt.configContent), 0600)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				defer os.Remove(configFile)
			}

			got, err := cfg.GetRegistryConfig()
			if err != nil {
				t.Fatalf("GetRegistryConfig() error = %v", err)
			}

			if got.URL != tt.want.URL {
				t.Errorf("GetRegistryConfig() URL = %v, want %v", got.URL, tt.want.URL)
			}
			if got.AuthType != tt.want.AuthType {
				t.Errorf("GetRegistryConfig() AuthType = %v, want %v", got.AuthType, tt.want.AuthType)
			}
		})
	}
}