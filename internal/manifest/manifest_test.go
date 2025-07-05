package manifest

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		manifestFile string
		content      string
		want         *Manifest
		wantErr      bool
	}{
		{
			name:         "valid manifest",
			manifestFile: filepath.Join(tempDir, "valid.toml"),
			content: `[package]
name = "test-package"
version = "1.2.3"
authors = ["Test Author <test@example.com>"]
description = "A test package"
license = "MIT"
repository = "https://github.com/test/package"
keywords = ["test", "package"]

[package.metadata]
main = "src/main.crl"
include = ["src/**/*.crl", "README.md"]
exclude = ["tests/**/*"]

[dependencies]
carrionlang = ">=0.1.6, <1.0.0"
json-utils = "^0.3.5"

[dev-dependencies]
test-framework = "~0.2.0"`,
			want: &Manifest{
				Package: Package{
					Name:        "test-package",
					Version:     "1.2.3",
					Authors:     []string{"Test Author <test@example.com>"},
					Description: "A test package",
					License:     "MIT",
					Repository:  "https://github.com/test/package",
					Keywords:    []string{"test", "package"},
					Metadata: PackageMetadata{
						Main:    "src/main.crl",
						Include: []string{"src/**/*.crl", "README.md"},
						Exclude: []string{"tests/**/*"},
					},
				},
				Dependencies: map[string]string{
					"carrionlang": ">=0.1.6, <1.0.0",
					"json-utils":  "^0.3.5",
				},
				DevDependencies: map[string]string{
					"test-framework": "~0.2.0",
				},
			},
		},
		{
			name:         "minimal manifest",
			manifestFile: filepath.Join(tempDir, "minimal.toml"),
			content: `[package]
name = "minimal"
version = "0.1.0"`,
			want: &Manifest{
				Package: Package{
					Name:    "minimal",
					Version: "0.1.0",
				},
			},
		},
		{
			name:         "manifest with arrays",
			manifestFile: filepath.Join(tempDir, "arrays.toml"),
			content: `[package]
name = "arrays-test"
version = "1.0.0"
authors = ["Author One", "Author Two", "Author Three"]
keywords = ["key1", "key2", "key3"]

[package.metadata]
include = ["src/**/*", "docs/**/*", "LICENSE", "README.md"]
exclude = ["*.log", "*.tmp", "build/**/*"]`,
			want: &Manifest{
				Package: Package{
					Name:     "arrays-test",
					Version:  "1.0.0",
					Authors:  []string{"Author One", "Author Two", "Author Three"},
					Keywords: []string{"key1", "key2", "key3"},
					Metadata: PackageMetadata{
						Include: []string{"src/**/*", "docs/**/*", "LICENSE", "README.md"},
						Exclude: []string{"*.log", "*.tmp", "build/**/*"},
					},
				},
			},
		},
		{
			name:         "empty dependencies",
			manifestFile: filepath.Join(tempDir, "empty-deps.toml"),
			content: `[package]
name = "no-deps"
version = "1.0.0"

[dependencies]

[dev-dependencies]`,
			want: &Manifest{
				Package: Package{
					Name:    "no-deps",
					Version: "1.0.0",
				},
				Dependencies:    map[string]string{},
				DevDependencies: map[string]string{},
			},
		},
		{
			name:         "invalid TOML",
			manifestFile: filepath.Join(tempDir, "invalid.toml"),
			content:      `[package invalid toml`,
			wantErr:      true,
		},
		{
			name:         "non-existent file",
			manifestFile: filepath.Join(tempDir, "nonexistent.toml"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != "" {
				err := os.WriteFile(tt.manifestFile, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			got, err := Load(tt.manifestFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !compareManifests(got, tt.want) {
					t.Errorf("Load() = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestWriteDefault(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "Bifrost.toml")

	err := WriteDefault(manifestPath)
	if err != nil {
		t.Fatalf("WriteDefault() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("manifest file was not created")
	}

	// Load and verify content
	m, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load written manifest: %v", err)
	}

	// Check default values
	if m.Package.Name != "my-package" {
		t.Errorf("default package name = %v, want my-package", m.Package.Name)
	}
	if m.Package.Version != "0.1.0" {
		t.Errorf("default version = %v, want 0.1.0", m.Package.Version)
	}
	if len(m.Package.Authors) != 1 || m.Package.Authors[0] != "Your Name <you@example.com>" {
		t.Errorf("default authors = %v, want [Your Name <you@example.com>]", m.Package.Authors)
	}
	if m.Package.Description != "A Carrion package" {
		t.Errorf("default description = %v, want 'A Carrion package'", m.Package.Description)
	}
	if m.Package.License != "MIT" {
		t.Errorf("default license = %v, want MIT", m.Package.License)
	}
	if m.Package.Metadata.Main != "src/main.crl" {
		t.Errorf("default main = %v, want src/main.crl", m.Package.Metadata.Main)
	}

	// Check include/exclude patterns
	expectedInclude := []string{"src/**/*.crl", "README.md", "LICENSE"}
	if !reflect.DeepEqual(m.Package.Metadata.Include, expectedInclude) {
		t.Errorf("default include = %v, want %v", m.Package.Metadata.Include, expectedInclude)
	}

	expectedExclude := []string{"tests/**/*", "*.log"}
	if !reflect.DeepEqual(m.Package.Metadata.Exclude, expectedExclude) {
		t.Errorf("default exclude = %v, want %v", m.Package.Metadata.Exclude, expectedExclude)
	}

	// Check empty maps
	if m.Dependencies == nil || len(m.Dependencies) != 0 {
		t.Errorf("default dependencies should be empty map, got %v", m.Dependencies)
	}
	if m.DevDependencies == nil || len(m.DevDependencies) != 0 {
		t.Errorf("default dev-dependencies should be empty map, got %v", m.DevDependencies)
	}
}

func TestWriteDefault_Error(t *testing.T) {
	// Test writing to an invalid path
	invalidPath := "/invalid/path/that/does/not/exist/Bifrost.toml"
	
	err := WriteDefault(invalidPath)
	if err == nil {
		t.Error("WriteDefault() should fail with invalid path")
	}
}

func TestManifestRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "roundtrip.toml")

	// Create a manifest with all fields populated
	original := &Manifest{
		Package: Package{
			Name:        "roundtrip-test",
			Version:     "2.5.8",
			Authors:     []string{"First Author", "Second Author"},
			Description: "Testing round-trip encoding/decoding",
			License:     "Apache-2.0",
			Repository:  "https://github.com/test/roundtrip",
			Keywords:    []string{"test", "roundtrip", "toml"},
			Metadata: PackageMetadata{
				Main:    "lib/index.crl",
				Include: []string{"lib/**/*.crl", "docs/**/*.md"},
				Exclude: []string{"**/*.test.crl", "tmp/**/*"},
			},
		},
		Dependencies: map[string]string{
			"dep1": "1.0.0",
			"dep2": "^2.0.0",
			"dep3": ">=3.0.0, <4.0.0",
		},
		DevDependencies: map[string]string{
			"dev1": "~1.2.3",
			"dev2": "0.5.0",
		},
	}

	// Write manifest
	err := WriteDefault(manifestPath)
	if err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Now write our test manifest
	content := `[package]
name = "roundtrip-test"
version = "2.5.8"
authors = ["First Author", "Second Author"]
description = "Testing round-trip encoding/decoding"
license = "Apache-2.0"
repository = "https://github.com/test/roundtrip"
keywords = ["test", "roundtrip", "toml"]

[package.metadata]
main = "lib/index.crl"
include = ["lib/**/*.crl", "docs/**/*.md"]
exclude = ["**/*.test.crl", "tmp/**/*"]

[dependencies]
dep1 = "1.0.0"
dep2 = "^2.0.0"
dep3 = ">=3.0.0, <4.0.0"

[dev-dependencies]
dev1 = "~1.2.3"
dev2 = "0.5.0"`

	err = os.WriteFile(manifestPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	// Load it back
	loaded, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	// Compare
	if !compareManifests(loaded, original) {
		t.Errorf("round-trip failed: loaded = %+v, want %+v", loaded, original)
	}
}

// Helper function to compare manifests
func compareManifests(a, b *Manifest) bool {
	if a == nil || b == nil {
		return a == b
	}

	// Compare Package fields
	if a.Package.Name != b.Package.Name ||
		a.Package.Version != b.Package.Version ||
		a.Package.Description != b.Package.Description ||
		a.Package.License != b.Package.License ||
		a.Package.Repository != b.Package.Repository ||
		a.Package.Metadata.Main != b.Package.Metadata.Main {
		return false
	}

	// Compare slices
	if !reflect.DeepEqual(a.Package.Authors, b.Package.Authors) ||
		!reflect.DeepEqual(a.Package.Keywords, b.Package.Keywords) ||
		!reflect.DeepEqual(a.Package.Metadata.Include, b.Package.Metadata.Include) ||
		!reflect.DeepEqual(a.Package.Metadata.Exclude, b.Package.Metadata.Exclude) {
		return false
	}

	// Compare maps
	if !compareMaps(a.Dependencies, b.Dependencies) ||
		!compareMaps(a.DevDependencies, b.DevDependencies) {
		return false
	}

	return true
}

func compareMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}