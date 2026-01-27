package manifest

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Manifest struct {
	Package         Package           `toml:"package"`
	Dependencies    map[string]string `toml:"dependencies"`
	DevDependencies map[string]string `toml:"dev-dependencies"`
}

type Package struct {
	Name        string          `toml:"name"`
	Version     string          `toml:"version"`
	Authors     []string        `toml:"authors"`
	Description string          `toml:"description"`
	License     string          `toml:"license"`
	Repository  string          `toml:"repository"`
	Keywords    []string        `toml:"keywords"`
	Metadata    PackageMetadata `toml:"metadata"`
}

type PackageMetadata struct {
	Main    string   `toml:"main"`
	Include []string `toml:"include"`
	Exclude []string `toml:"exclude"`
}

func Load(path string) (*Manifest, error) {
	var m Manifest
	if _, err := toml.DecodeFile(path, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func WriteDefault(path string, packageName string, versionNumber string) error {
	if packageName == "" {
		packageName = "default-package"
	}
	if versionNumber == "" {
		versionNumber = "0.0.1"
	}
	m := Manifest{
		Package: Package{
			Name:        packageName,
			Version:     versionNumber,
			Authors:     []string{"Your Name <you@example.com>"},
			Description: "A Carrion package",
			License:     "MIT",
			Repository:  "",
			Keywords:    []string{},
			Metadata: PackageMetadata{
				Main:    "src/main.crl",
				Include: []string{"src/**/*.crl", "README.md", "LICENSE"},
				Exclude: []string{"appraise/**/*", "*.log"},
			},
		},
		Dependencies:    map[string]string{},
		DevDependencies: map[string]string{},
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(m)
}
