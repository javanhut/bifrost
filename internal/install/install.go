package install

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/javanhut/bifrost/internal/config"
	"github.com/javanhut/bifrost/internal/manifest"
	"github.com/javanhut/bifrost/internal/resolver"
)

type Installer struct {
	config *config.Config
}

func New(cfg *config.Config) *Installer {
	return &Installer{
		config: cfg,
	}
}

func (i *Installer) Install(resolution *resolver.Resolution) error {
	// Get installation order
	packages := resolution.GetResolutionOrder()

	for _, pkg := range packages {
		fmt.Printf("Installing %s@%s...\n", pkg.Name, pkg.Version)
		
		if err := i.installPackage(pkg); err != nil {
			return fmt.Errorf("failed to install %s: %w", pkg.Name, err)
		}
	}

	return nil
}

func (i *Installer) installPackage(pkg *resolver.Package) error {
	// Check if already installed
	installPath := i.config.PackagePath(pkg.Name, pkg.Version.String())
	if _, err := os.Stat(installPath); err == nil {
		fmt.Printf("  Already installed at %s\n", installPath)
		return nil
	}

	// TODO: Download from registry when available
	// For now, we'll implement local file installation
	
	return fmt.Errorf("registry download not yet implemented")
}

func (i *Installer) InstallLocal(manifestPath string) error {
	// Load manifest
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Get package directory
	packageDir := filepath.Dir(manifestPath)
	
	// Create local modules directory
	modulesDir := filepath.Join(packageDir, i.config.LocalModulesPath())
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create modules directory: %w", err)
	}
	
	// This is a no-op for local development
	fmt.Printf("Local package %s is ready for development\n", m.Package.Name)
	fmt.Printf("Import path: %s\n", m.Package.Name)
	
	return nil
}

func (i *Installer) InstallFromArchive(archivePath string, pkg *resolver.Package) error {
	// Create target directory
	installPath := i.config.PackagePath(pkg.Name, pkg.Version.String())
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	// Open archive
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Extract based on file extension
	if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		return i.extractTarGz(file, installPath)
	}

	return fmt.Errorf("unsupported archive format")
}

func (i *Installer) extractTarGz(r io.Reader, destDir string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create directory if needed
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			// Create file
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// Copy contents
			if _, err := io.Copy(file, tr); err != nil {
				file.Close()
				return err
			}
			file.Close()
		}
	}

	return nil
}

func (i *Installer) Download(url string, destPath string) error {
	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Create file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy contents
	_, err = io.Copy(out, resp.Body)
	return err
}

func (i *Installer) CreateSymlinks(pkg *resolver.Package) error {
	// Create symlink in local modules directory for easy imports
	installPath := i.config.PackagePath(pkg.Name, pkg.Version.String())
	linkPath := filepath.Join(i.config.LocalModulesPath(), pkg.Name)

	// Remove existing link if present
	os.Remove(linkPath)

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
		return err
	}

	// Create symlink
	return os.Symlink(installPath, linkPath)
}