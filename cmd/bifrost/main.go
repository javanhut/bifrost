package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/javanhut/bifrost/internal/config"
	"github.com/javanhut/bifrost/internal/install"
	"github.com/javanhut/bifrost/internal/manifest"
	"github.com/spf13/cobra"
)

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
				cmd.Printf("Installing %s...\n", args[0])
				if global {
					cmd.Println("Global installation not yet implemented (requires registry)")
				} else {
					cmd.Println("Registry integration not yet implemented")
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

	// Search command (placeholder)
	root.AddCommand(&cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Searching for '%s'...\n", args[0])
			cmd.Println("Registry integration not yet implemented")
		},
	})

	// Info command
	root.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Show package information",
		Run: func(cmd *cobra.Command, args []string) {
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
		},
	})

	// Version command
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show Bifrost version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Bifrost 0.1.0")
			cmd.Println("Package manager for the Carrion programming language")
		},
	})

	cobra.CheckErr(root.Execute())
}
