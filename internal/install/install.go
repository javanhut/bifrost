package install

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/javanhut/bifrost/internal/config"
	"github.com/javanhut/bifrost/internal/manifest"
	"github.com/javanhut/bifrost/internal/registry"
	"github.com/javanhut/bifrost/internal/resolver"
	ver "github.com/javanhut/bifrost/internal/version"
)

type Installer struct {
	config *config.Config
}

// getAPIURL extracts the API URL from the registry URL
func getAPIURL(registryURL string) string {
	u, err := url.Parse(registryURL)
	if err != nil {
		return registryURL
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
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
	// Check if already installed (user-specific location)
	installPath := i.config.PackagePath(pkg.Name, pkg.Version.String())
	if _, err := os.Stat(installPath); err == nil {
		fmt.Printf("  Already installed at %s\n", installPath)
		return nil
	}

	// Download from registry
	client := registry.NewClient(i.config.RegistryURL)

	// Get package info to get download URL
	_, err := client.GetPackageInfo(pkg.Name, pkg.Version.String())
	if err != nil {
		return fmt.Errorf("failed to get package info: %w", err)
	}

	// Download package archive
	archivePath := i.config.CachePath(fmt.Sprintf("%s-%s.tar.gz", pkg.Name, pkg.Version.String()))
	
	fmt.Printf("  Downloading %s@%s...\n", pkg.Name, pkg.Version.String())
	reader, err := client.DownloadPackage(pkg.Name, pkg.Version.String())
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer reader.Close()
	
	// Save to file
	if err := i.saveToFile(reader, archivePath); err != nil {
		return fmt.Errorf("failed to save package: %w", err)
	}

	// Install from archive
	fmt.Printf("  Extracting to %s...\n", installPath)
	if err := i.InstallFromArchive(archivePath, pkg); err != nil {
		return fmt.Errorf("failed to install from archive: %w", err)
	}

	// Clean up archive after successful installation
	os.Remove(archivePath)

	fmt.Printf("  Successfully installed %s@%s\n", pkg.Name, pkg.Version.String())
	return nil
}

// InstallGlobal installs a package to the shared global location
func (i *Installer) InstallGlobal(pkg *resolver.Package, sourcePath string) error {
	sharedDir := i.config.GetSharedGlobalPackagesDir()
	installPath := filepath.Join(sharedDir, pkg.Name, pkg.Version.String())

	// Check if already installed globally
	if _, err := os.Stat(installPath); err == nil {
		fmt.Printf("Package %s@%s already installed globally at %s\n",
			pkg.Name, pkg.Version.String(), installPath)
		return nil
	}

	// Create target directory (may require sudo)
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create global install directory %s (may need sudo): %w",
			installPath, err)
	}

	// Copy package files to global location
	if err := i.copyDirectory(sourcePath, installPath); err != nil {
		return fmt.Errorf("failed to copy package to global location: %w", err)
	}

	fmt.Printf("Package %s@%s installed globally at %s\n",
		pkg.Name, pkg.Version.String(), installPath)

	return nil
}

// copyDirectory recursively copies a directory
func (i *Installer) copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return i.copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func (i *Installer) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
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

func (i *Installer) saveToFile(reader io.Reader, destPath string) error {
	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Create file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy contents
	_, err = io.Copy(out, reader)
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

func (i *Installer) InstallPackageByName(packageName string, version string, global bool) error {
	// For local installation, use the local package installation method
	if !global {
		return i.InstallPackageLocalByName(packageName, version)
	}
	
	// Continue with existing global installation logic
	client := registry.NewClient(i.config.RegistryURL)

	// If no version specified, get latest
	if version == "" {
		version = "latest"
	}

	// Get package info
	pkgInfo, err := client.GetPackageInfo(packageName, version)
	if err != nil {
		return fmt.Errorf("failed to get package info: %w", err)
	}

	// Create a Package struct for installation
	pkg := &resolver.Package{
		Name: pkgInfo.Name,
		Version: &ver.Version{
			Major: 0,
			Minor: 0,
			Patch: 1,
		},
	}

	// Parse actual version if not "latest"
	if version != "latest" && pkgInfo.Version != "" {
		if v, err := ver.Parse(pkgInfo.Version); err == nil {
			pkg.Version = v
		}
	}

	if global {
		// For global install, we need to download first then install globally
		archivePath := i.config.CachePath(fmt.Sprintf("%s-%s.tar.gz", pkg.Name, pkg.Version.String()))

		fmt.Printf("Downloading %s@%s...\n", pkg.Name, pkgInfo.Version)
		reader, err := client.DownloadPackage(pkg.Name, pkgInfo.Version)
		if err != nil {
			return fmt.Errorf("failed to download package: %w", err)
		}
		defer reader.Close()
		
		// Save to file
		if err := i.saveToFile(reader, archivePath); err != nil {
			return fmt.Errorf("failed to save package: %w", err)
		}

		// Extract to temp location
		tempDir := i.config.CachePath(fmt.Sprintf("%s-%s-temp", pkg.Name, pkg.Version.String()))
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		// Open archive file
		archiveFile, err := os.Open(archivePath)
		if err != nil {
			return fmt.Errorf("failed to open archive: %w", err)
		}
		defer archiveFile.Close()

		if err := i.extractTarGz(archiveFile, tempDir); err != nil {
			return fmt.Errorf("failed to extract package: %w", err)
		}

		// Install globally
		if err := i.InstallGlobal(pkg, tempDir); err != nil {
			return err
		}

		// Clean up
		os.Remove(archivePath)
	} else {
		// Regular user-specific install
		if err := i.installPackage(pkg); err != nil {
			return err
		}
	}

	return nil
}

// InstallPackageLocalByName installs a package to the local project directory
func (i *Installer) InstallPackageLocalByName(packageName string, version string) error {
	client := registry.NewClient(i.config.RegistryURL)

	// If no version specified, get latest
	if version == "" {
		version = "latest"
	}

	// Get package info
	pkgInfo, err := client.GetPackageInfo(packageName, version)
	if err != nil {
		return fmt.Errorf("failed to get package info: %w", err)
	}

	// Create a Package struct for installation
	pkg := &resolver.Package{
		Name: pkgInfo.Name,
		Version: &ver.Version{
			Major: 0,
			Minor: 0,
			Patch: 1,
		},
	}

	// Parse actual version if not "latest"
	if version != "latest" && pkgInfo.Version != "" {
		if v, err := ver.Parse(pkgInfo.Version); err == nil {
			pkg.Version = v
		}
	}

	// Use local installation path
	installPath := i.config.LocalPackagePath(pkg.Name, pkgInfo.Version)
	
	// Check if already installed locally
	if _, err := os.Stat(installPath); err == nil {
		fmt.Printf("Package %s@%s already installed locally at %s\n", pkg.Name, pkgInfo.Version, installPath)
		return nil
	}

	// Download package archive
	archivePath := i.config.CachePath(fmt.Sprintf("%s-%s.tar.gz", pkg.Name, pkgInfo.Version))
	
	fmt.Printf("Downloading %s@%s...\n", pkg.Name, pkgInfo.Version)
	reader, err := client.DownloadPackage(pkg.Name, pkgInfo.Version)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer reader.Close()
	
	// Save to file
	if err := i.saveToFile(reader, archivePath); err != nil {
		return fmt.Errorf("failed to save package: %w", err)
	}

	// Install from archive to local directory
	fmt.Printf("Installing to %s...\n", installPath)
	if err := i.InstallFromArchiveToLocal(archivePath, pkg, pkgInfo.Version); err != nil {
		return fmt.Errorf("failed to install from archive: %w", err)
	}

	// Clean up archive after successful installation
	os.Remove(archivePath)

	fmt.Printf("Successfully installed %s@%s to %s\n", pkg.Name, pkgInfo.Version, installPath)
	return nil
}

// InstallFromArchiveToLocal extracts an archive to a local project directory
func (i *Installer) InstallFromArchiveToLocal(archivePath string, pkg *resolver.Package, version string) error {
	// Create target directory in local modules
	installPath := i.config.LocalPackagePath(pkg.Name, version)
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
