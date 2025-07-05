package uninstall

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/javanhut/bifrost/internal/config"
	"github.com/javanhut/bifrost/internal/manifest"
)

type Uninstaller struct {
	config *config.Config
}

func New(cfg *config.Config) *Uninstaller {
	return &Uninstaller{
		config: cfg,
	}
}

func (u *Uninstaller) UninstallPackage(packageName string, version string, global bool) error {
	if version == "" {
		return u.uninstallAllVersions(packageName, global)
	}
	return u.uninstallSpecificVersion(packageName, version, global)
}

func (u *Uninstaller) uninstallSpecificVersion(packageName string, version string, global bool) error {
	var packagePath string
	
	if global {
		sharedDir := u.config.GetSharedGlobalPackagesDir()
		packagePath = filepath.Join(sharedDir, packageName, version)
	} else {
		// Check local carrion_modules first, then user packages
		localPath := u.config.LocalPackagePath(packageName, version)
		if _, err := os.Stat(localPath); err == nil {
			packagePath = localPath
		} else {
			packagePath = u.config.PackagePath(packageName, version)
		}
	}

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		installType := "locally"
		if global {
			installType = "globally"
		}
		return fmt.Errorf("package %s@%s is not installed %s", packageName, version, installType)
	}

	fmt.Printf("Removing %s@%s", packageName, version)
	if global {
		fmt.Print(" (global)")
	}
	fmt.Println("...")

	if err := os.RemoveAll(packagePath); err != nil {
		return fmt.Errorf("failed to remove package directory: %w", err)
	}

	u.cleanupSymlinks(packageName)

	packageDir := filepath.Dir(packagePath)
	if isEmpty, _ := u.isDirEmpty(packageDir); isEmpty {
		os.Remove(packageDir)
	}

	fmt.Printf("Successfully removed %s@%s", packageName, version)
	if global {
		fmt.Print(" (global)")
	}
	fmt.Println()

	return nil
}

func (u *Uninstaller) uninstallAllVersions(packageName string, global bool) error {
	var packageDir string
	
	if global {
		sharedDir := u.config.GetSharedGlobalPackagesDir()
		packageDir = filepath.Join(sharedDir, packageName)
	} else {
		// Check local carrion_modules first, then user packages
		localPackageDir := filepath.Join(u.config.ModulesDir, packageName)
		if _, err := os.Stat(localPackageDir); err == nil {
			packageDir = localPackageDir
		} else {
			packageDir = filepath.Join(u.config.PackagesDir, packageName)
		}
	}

	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		installType := "locally"
		if global {
			installType = "globally"
		}
		return fmt.Errorf("package %s is not installed %s", packageName, installType)
	}

	versions, err := os.ReadDir(packageDir)
	if err != nil {
		return fmt.Errorf("failed to read package directory: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no versions found for package %s", packageName)
	}

	fmt.Printf("Removing all versions of %s", packageName)
	if global {
		fmt.Print(" (global)")
	}
	fmt.Println("...")

	for _, version := range versions {
		if version.IsDir() {
			fmt.Printf("  Removing %s@%s...\n", packageName, version.Name())
		}
	}

	if err := os.RemoveAll(packageDir); err != nil {
		return fmt.Errorf("failed to remove package directory: %w", err)
	}

	u.cleanupSymlinks(packageName)

	fmt.Printf("Successfully removed all versions of %s", packageName)
	if global {
		fmt.Print(" (global)")
	}
	fmt.Println()

	return nil
}

func (u *Uninstaller) cleanupSymlinks(packageName string) {
	linkPath := filepath.Join(u.config.LocalModulesPath(), packageName)
	if _, err := os.Lstat(linkPath); err == nil {
		if isSymlink, _ := u.isSymlink(linkPath); isSymlink {
			os.Remove(linkPath)
		}
	}
}

func (u *Uninstaller) isSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}

func (u *Uninstaller) isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func (u *Uninstaller) ListInstalledPackages(global bool) error {
	if global {
		return u.listPackagesInDir(u.config.GetSharedGlobalPackagesDir(), "Global")
	} else {
		// List both local project and user packages for non-global mode
		fmt.Println("Installed packages:")
		
		// Check local carrion_modules first
		localFound := false
		if _, err := os.Stat(u.config.ModulesDir); err == nil {
			fmt.Println("\nLocal project packages (./carrion_modules/):")
			if err := u.listPackagesInDir(u.config.ModulesDir, ""); err == nil {
				localFound = true
			}
		}
		
		// Then check user packages
		userFound := false
		if _, err := os.Stat(u.config.PackagesDir); err == nil {
			fmt.Println("\nUser packages (~/.carrion/packages/):")
			if err := u.listPackagesInDir(u.config.PackagesDir, ""); err == nil {
				userFound = true
			}
		}
		
		if !localFound && !userFound {
			fmt.Println("No packages installed")
		}
		
		return nil
	}
}

func (u *Uninstaller) listPackagesInDir(packagesDir string, installType string) error {
	entries, err := os.ReadDir(packagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			if installType != "" {
				fmt.Printf("No %s packages installed\n", strings.ToLower(installType))
			}
			return err
		}
		return fmt.Errorf("failed to read packages directory: %w", err)
	}

	if len(entries) == 0 {
		if installType != "" {
			fmt.Printf("No %s packages installed\n", strings.ToLower(installType))
		}
		return err
	}

	if installType != "" {
		fmt.Printf("%s packages:\n", installType)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			versionsDir := filepath.Join(packagesDir, entry.Name())
			versions, err := os.ReadDir(versionsDir)
			if err != nil {
				continue
			}
			for _, version := range versions {
				if version.IsDir() {
					fmt.Printf("  %s@%s\n", entry.Name(), version.Name())
				}
			}
		}
	}

	return nil
}

func (u *Uninstaller) UninstallFromManifest(manifestPath string) error {
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	fmt.Printf("Removing dependencies for %s...\n", m.Package.Name)

	allDeps := make(map[string]string)
	for name, version := range m.Dependencies {
		allDeps[name] = version
	}
	for name, version := range m.DevDependencies {
		allDeps[name] = version
	}

	if len(allDeps) == 0 {
		fmt.Println("No dependencies to remove")
		return nil
	}

	for name, version := range allDeps {
		if strings.Contains(version, "*") || strings.Contains(version, "^") || strings.Contains(version, "~") {
			if err := u.uninstallAllVersions(name, false); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", name, err)
				continue
			}
		} else {
			if err := u.uninstallSpecificVersion(name, version, false); err != nil {
				fmt.Printf("Warning: failed to remove %s@%s: %v\n", name, version, err)
				continue
			}
		}
	}

	modulesDir := filepath.Join(filepath.Dir(manifestPath), u.config.LocalModulesPath())
	if _, err := os.Stat(modulesDir); err == nil {
		if isEmpty, _ := u.isDirEmpty(modulesDir); isEmpty {
			os.Remove(modulesDir)
			fmt.Printf("Removed empty %s directory\n", u.config.LocalModulesPath())
		}
	}

	fmt.Printf("Finished removing dependencies for %s\n", m.Package.Name)
	return nil
}

func (u *Uninstaller) CleanCache() error {
	fmt.Println("Cleaning package cache...")
	
	if _, err := os.Stat(u.config.CacheDir); os.IsNotExist(err) {
		fmt.Println("Cache directory does not exist")
		return nil
	}

	entries, err := os.ReadDir(u.config.CacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("Cache is already empty")
		return nil
	}

	for _, entry := range entries {
		cachePath := filepath.Join(u.config.CacheDir, entry.Name())
		if err := os.RemoveAll(cachePath); err != nil {
			fmt.Printf("Warning: failed to remove %s: %v\n", entry.Name(), err)
			continue
		}
		fmt.Printf("  Removed %s\n", entry.Name())
	}

	fmt.Printf("Successfully cleaned cache (%d items removed)\n", len(entries))
	return nil
}