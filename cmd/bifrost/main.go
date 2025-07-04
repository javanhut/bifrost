package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/javanhut/bifrost/internal/config"
	"github.com/javanhut/bifrost/internal/install"
	"github.com/javanhut/bifrost/internal/manifest"
	"github.com/javanhut/bifrost/internal/registry"
	"github.com/spf13/cobra"
)

func runCommand(cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	// Initialize configuration
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Ensure directories exist
	if err := cfg.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	root := &cobra.Command{
		Use:   "bifrost",
		Short: "Bifrost - Carrion's package manager",
		Long:  "Bifrost is the package manager for the Carrion programming language",
	}

	// Init command
	root.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create a new Carrion package",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat("Bifrost.toml"); err == nil {
				cmd.PrintErrln("Error: Bifrost.toml already exists")
				os.Exit(1)
			}

			err := manifest.WriteDefault("Bifrost.toml")
			cobra.CheckErr(err)

			// Create directory structure
			dirs := []string{"src", "tests", "docs"}
			for _, dir := range dirs {
				os.MkdirAll(dir, 0755)
			}

			// Create main.crl
			mainContent := `# Main module for your Carrion package

grimoire Main {
    incantation new() {
        echo("Hello from your new Carrion package!")
    }
}
`
			os.WriteFile("src/main.crl", []byte(mainContent), 0644)

			cmd.Println("Created new Carrion package")
			cmd.Println("Edit Bifrost.toml to configure your package")
		},
	})

	// Install command
	installCmd := &cobra.Command{
		Use:   "install [package]",
		Short: "Install dependencies",
		Run: func(cmd *cobra.Command, args []string) {
			installer := install.New(cfg)
			global, _ := cmd.Flags().GetBool("global")

			if len(args) == 0 {
				// Install from Bifrost.toml
				_, err := manifest.Load("Bifrost.toml")
				if err != nil {
					cmd.PrintErrf("Error loading Bifrost.toml: %v\n", err)
					os.Exit(1)
				}

				// TODO: Load available packages from registry
				cmd.Println("Installing dependencies from Bifrost.toml...")
				cmd.Println("Registry integration not yet implemented")

				// For now, just install local package
				err = installer.InstallLocal("Bifrost.toml")
				cobra.CheckErr(err)
			} else {
				// Install specific package
				packageName := args[0]
				version := ""

				// Parse package@version format
				if idx := strings.Index(packageName, "@"); idx != -1 {
					version = packageName[idx+1:]
					packageName = packageName[:idx]
				}

				cmd.Printf("Installing %s", packageName)
				if version != "" {
					cmd.Printf("@%s", version)
				}
				cmd.Println("...")

				err := installer.InstallPackageByName(packageName, version, global)
				if err != nil {
					cmd.PrintErrf("Error installing package: %v\n", err)
					os.Exit(1)
				}
			}
		},
	}
	installCmd.Flags().BoolP("global", "g", false, "Install package globally")
	root.AddCommand(installCmd)

	// List command
	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		Run: func(cmd *cobra.Command, args []string) {
			// List packages in global directory
			packagesDir := cfg.PackagesDir
			entries, err := os.ReadDir(packagesDir)
			if err != nil {
				if os.IsNotExist(err) {
					cmd.Println("No packages installed")
					return
				}
				cobra.CheckErr(err)
			}

			if len(entries) == 0 {
				cmd.Println("No packages installed")
				return
			}

			cmd.Println("Installed packages:")
			for _, entry := range entries {
				if entry.IsDir() {
					// List versions
					versionsDir := filepath.Join(packagesDir, entry.Name())
					versions, err := os.ReadDir(versionsDir)
					if err != nil {
						continue
					}
					for _, version := range versions {
						if version.IsDir() {
							cmd.Printf("  %s@%s\n", entry.Name(), version.Name())
						}
					}
				}
			}
		},
	})

	// Search command
	root.AddCommand(&cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := registry.NewClient(cfg.RegistryURL)

			// Check registry health first
			if err := client.Health(); err != nil {
				cmd.PrintErrf("Error connecting to registry: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("Searching for '%s'...\n", args[0])
			results, err := client.Search(args[0])
			if err != nil {
				cmd.PrintErrf("Error searching packages: %v\n", err)
				os.Exit(1)
			}

			if len(results) == 0 {
				cmd.Println("No packages found.")
				return
			}

			cmd.Printf("Found %d package(s):\n\n", len(results))
			for _, pkg := range results {
				cmd.Printf("  %s (%s)\n", pkg.Name, pkg.Version)
				if pkg.Description != "" {
					cmd.Printf("    %s\n", pkg.Description)
				}
				if pkg.Downloads > 0 {
					cmd.Printf("    Downloads: %d\n", pkg.Downloads)
				}
				cmd.Println()
			}
		},
	})

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info [package]",
		Short: "Show package information",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				// Show local package info
				m, err := manifest.Load("Bifrost.toml")
				if err != nil {
					cmd.PrintErrf("Error loading Bifrost.toml: %v\n", err)
					os.Exit(1)
				}

				cmd.Printf("Package: %s\n", m.Package.Name)
				cmd.Printf("Version: %s\n", m.Package.Version)
				cmd.Printf("Description: %s\n", m.Package.Description)
				if len(m.Package.Authors) > 0 {
					cmd.Printf("Authors: %v\n", m.Package.Authors)
				}
				if m.Package.License != "" {
					cmd.Printf("License: %s\n", m.Package.License)
				}

				if len(m.Dependencies) > 0 {
					cmd.Println("\nDependencies:")
					for name, version := range m.Dependencies {
						cmd.Printf("  %s: %s\n", name, version)
					}
				}

				if len(m.DevDependencies) > 0 {
					cmd.Println("\nDev Dependencies:")
					for name, version := range m.DevDependencies {
						cmd.Printf("  %s: %s\n", name, version)
					}
				}
			} else {
				// Show remote package info
				packageName := args[0]
				version := ""

				// Parse package@version format
				if idx := strings.Index(packageName, "@"); idx != -1 {
					version = packageName[idx+1:]
					packageName = packageName[:idx]
				}

				client := registry.NewClient(cfg.RegistryURL)

				var pkgInfo *registry.PackageInfo
				var err error

				if version == "" {
					pkgInfo, err = client.GetPackageLatest(packageName)
				} else {
					pkgInfo, err = client.GetPackageInfo(packageName, version)
				}

				if err != nil {
					cmd.PrintErrf("Error fetching package info: %v\n", err)
					os.Exit(1)
				}

				cmd.Printf("Package: %s\n", pkgInfo.Name)
				cmd.Printf("Version: %s\n", pkgInfo.Version)
				if pkgInfo.Description != "" {
					cmd.Printf("Description: %s\n", pkgInfo.Description)
				}
				if len(pkgInfo.Authors) > 0 {
					cmd.Printf("Authors: %v\n", pkgInfo.Authors)
				}
				if pkgInfo.License != "" {
					cmd.Printf("License: %s\n", pkgInfo.License)
				}
				if pkgInfo.Homepage != "" {
					cmd.Printf("Homepage: %s\n", pkgInfo.Homepage)
				}
				if pkgInfo.Repository != "" {
					cmd.Printf("Repository: %s\n", pkgInfo.Repository)
				}
				if len(pkgInfo.Keywords) > 0 {
					cmd.Printf("Keywords: %v\n", pkgInfo.Keywords)
				}
			}
		},
	}
	root.AddCommand(infoCmd)

	// Publish command
	publishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a package to the registry",
		Run: func(cmd *cobra.Command, args []string) {
			// Load manifest
			m, err := manifest.Load("Bifrost.toml")
			if err != nil {
				cmd.PrintErrf("Error loading Bifrost.toml: %v\n", err)
				os.Exit(1)
			}

			// Create archive
			archiveName := fmt.Sprintf("%s-%s.tar.gz", m.Package.Name, m.Package.Version)
			archivePath := filepath.Join(os.TempDir(), archiveName)

			cmd.Printf("Creating package archive %s...\n", archiveName)

			// Create archive using tar command
			tarCmd := fmt.Sprintf("tar -czf %s --exclude='.git' --exclude='*.tar.gz' --exclude='bifrost' --exclude='carrion_modules' .", archivePath)
			if err := runCommand(tarCmd); err != nil {
				cmd.PrintErrf("Error creating archive: %v\n", err)
				os.Exit(1)
			}
			defer os.Remove(archivePath)

			// Prepare metadata
			metadata := &registry.PackageInfo{
				Name:        m.Package.Name,
				Version:     m.Package.Version,
				Description: m.Package.Description,
				Authors:     m.Package.Authors,
				License:     m.Package.License,
				Homepage:    "", // Not in manifest yet
				Repository:  m.Package.Repository,
				Keywords:    m.Package.Keywords,
			}

			// Publish to registry
			client := registry.NewClient(cfg.RegistryURL)

			cmd.Printf("Publishing %s@%s to %s...\n", m.Package.Name, m.Package.Version, cfg.RegistryURL)
			if err := client.Publish(archivePath, metadata); err != nil {
				cmd.PrintErrf("Error publishing package: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("Successfully published %s@%s!\n", m.Package.Name, m.Package.Version)
		},
	}
	root.AddCommand(publishCmd)

	// Version command
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show Bifrost version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Bifrost 0.2.0")
			cmd.Println("Package manager for the Carrion programming language")
			cmd.Printf("Registry: %s\n", cfg.RegistryURL)
		},
	})

	cobra.CheckErr(root.Execute())
}
