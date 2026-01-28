package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"errors"
	"regexp"

	"github.com/javanhut/bifrost/internal/auth"
	"github.com/javanhut/bifrost/internal/config"
	"github.com/javanhut/bifrost/internal/install"
	"github.com/javanhut/bifrost/internal/manifest"
	"github.com/javanhut/bifrost/internal/registry"
	"github.com/javanhut/bifrost/internal/uninstall"
	"github.com/spf13/cobra"
)

// embeddedManifest will be populated at build time if using go:embed
// For now, we'll rely on runtime loading
var embeddedManifest string

func runCommand(cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
func validateVersion(s string) error {
				var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
				if !versionRegex.MatchString(s){
					return errors.New("Invalid format must match format: 0.0.0 or Major.Minor.Patch")
				}
				return nil
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
			var packageName string
			var versionNumber string
			if len(args) == 1 {	
				packageName = args[0]
			}
			if len(args) == 2 {
				packageName = args[0]
				versionNumber = args[1]
				err := validateVersion(versionNumber)
				if err != nil {
					errorStr := fmt.Sprintf("Error: %v", err)
					cmd.PrintErrln(errorStr)
				}


			}
			if _, err := os.Stat("Bifrost.toml"); err == nil {
				cmd.PrintErrln("Error: Bifrost.toml already exists")
				os.Exit(1)
			}
			packageName = strings.ToLower(packageName)
			makePackage := packageName
			os.MkdirAll(makePackage, 0755)
			// Create directory structure
			tomlPath := fmt.Sprintf("%s/Bifrost.toml", packageName)
			err := manifest.WriteDefault(tomlPath, packageName, versionNumber)
			cobra.CheckErr(err)
			dirs := []string{"src", "appraise", "docs"}
			for _, dir := range dirs {
				insidePackage := fmt.Sprintf("%s/%s",packageName, dir)
				os.MkdirAll(insidePackage, 0755)
			}

			// Create main.crl
			mainContent := fmt.Sprintf(`grim Main:
    init():
        self.name = "%s"
    spell new():
        return self.name
main:
    m = Main()
    m.new()
`, packageName)
			testContent := fmt.Sprintf(`import "src/main"

spell appraise_main():
    m = Main()
    check(m.new() == "%s")
`, packageName)
			readmeInfo := fmt.Sprintf("# %s", packageName)
			mainFile := fmt.Sprintf("%s/src/main.crl",packageName)
			testFile := fmt.Sprintf("%s/appraise/appraise_main.crl",packageName)
			docsFile := fmt.Sprintf("%s/docs/README.md", packageName)
			os.WriteFile(mainFile, []byte(mainContent), 0644)
			os.WriteFile(testFile, []byte(testContent),0644)
			os.WriteFile(docsFile, []byte(readmeInfo), 0644)
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

	// Uninstall command
	uninstallCmd := &cobra.Command{
		Use:   "uninstall [package]",
		Short: "Uninstall packages",
		Run: func(cmd *cobra.Command, args []string) {
			uninstaller := uninstall.New(cfg)
			global, _ := cmd.Flags().GetBool("global")
			all, _ := cmd.Flags().GetBool("all")
			clean, _ := cmd.Flags().GetBool("clean")

			if clean {
				err := uninstaller.CleanCache()
				if err != nil {
					cmd.PrintErrf("Error cleaning cache: %v\n", err)
					os.Exit(1)
				}
				return
			}

			if len(args) == 0 {
				if all {
					cmd.PrintErrln("Error: --all flag requires specifying a package name")
					os.Exit(1)
				}
				
				// Uninstall from Bifrost.toml
				_, err := manifest.Load("Bifrost.toml")
				if err != nil {
					cmd.PrintErrf("Error loading Bifrost.toml: %v\n", err)
					os.Exit(1)
				}

				cmd.Println("Uninstalling dependencies from Bifrost.toml...")
				err = uninstaller.UninstallFromManifest("Bifrost.toml")
				if err != nil {
					cmd.PrintErrf("Error uninstalling dependencies: %v\n", err)
					os.Exit(1)
				}
			} else {
				// Uninstall specific package
				packageName := args[0]
				version := ""

				if !all {
					// Parse package@version format
					if idx := strings.Index(packageName, "@"); idx != -1 {
						version = packageName[idx+1:]
						packageName = packageName[:idx]
					}
				}

				err := uninstaller.UninstallPackage(packageName, version, global)
				if err != nil {
					cmd.PrintErrf("Error uninstalling package: %v\n", err)
					os.Exit(1)
				}
			}
		},
	}
	uninstallCmd.Flags().BoolP("global", "g", false, "Uninstall package globally")
	uninstallCmd.Flags().BoolP("all", "a", false, "Uninstall all versions of the package")
	uninstallCmd.Flags().BoolP("clean", "c", false, "Clean package cache")
	root.AddCommand(uninstallCmd)

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		Run: func(cmd *cobra.Command, args []string) {
			global, _ := cmd.Flags().GetBool("global")
			uninstaller := uninstall.New(cfg)

			if global {
				err := uninstaller.ListInstalledPackages(true)
				if err != nil {
					cmd.PrintErrf("Error listing global packages: %v\n", err)
					os.Exit(1)
				}
			} else {
				err := uninstaller.ListInstalledPackages(false)
				if err != nil {
					cmd.PrintErrf("Error listing local packages: %v\n", err)
					os.Exit(1)
				}
			}
		},
	}
	listCmd.Flags().BoolP("global", "g", false, "List globally installed packages")
	root.AddCommand(listCmd)

	// Search command
	root.AddCommand(&cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			registryConfig, err := cfg.GetRegistryConfig()
			if err != nil {
				cmd.PrintErrf("Error loading registry config: %v\n", err)
				os.Exit(1)
			}
			
			client := registry.NewClient(registryConfig.URL)

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

				registryConfig, err := cfg.GetRegistryConfig()
				if err != nil {
					cmd.PrintErrf("Error loading registry config: %v\n", err)
					os.Exit(1)
				}
				
				client := registry.NewClient(registryConfig.URL)

				var pkgInfo *registry.PackageInfo

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

	// Login command
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the package registry",
		Long:  "Authenticate with the package registry to enable publishing packages",
		Run: func(cmd *cobra.Command, args []string) {
			authService := auth.New(cfg)
			
			if err := authService.Login(); err != nil {
				cmd.PrintErrf("Login failed: %v\n", err)
				os.Exit(1)
			}
		},
	}
	root.AddCommand(loginCmd)

	// Logout command
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove authentication credentials",
		Long:  "Remove stored authentication credentials for the package registry",
		Run: func(cmd *cobra.Command, args []string) {
			authService := auth.New(cfg)
			
			if err := authService.Logout(); err != nil {
				cmd.PrintErrf("Logout failed: %v\n", err)
				os.Exit(1)
			}
		},
	}
	root.AddCommand(logoutCmd)

	// Publish command
	publishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a package to the registry",
		Run: func(cmd *cobra.Command, args []string) {
			// Get registry configuration (stored config + environment overrides)
			registryConfig, err := cfg.GetRegistryConfig()
			if err != nil {
				cmd.PrintErrf("Error loading registry config: %v\n", err)
				os.Exit(1)
			}

			// Check if we have authentication configured
			if registryConfig.AuthType == "none" || registryConfig.AuthType == "" {
				// Try to use legacy auth file
				authService := auth.New(cfg)
				authConfig, err := authService.GetAuthConfig()
				if err != nil {
					cmd.PrintErrf("Authentication required: %v\n", err)
					cmd.PrintErrln("Configure credentials with 'bifrost config set registry.username <user>' and 'bifrost config set registry.password <pass>'")
					cmd.PrintErrln("Or run 'bifrost login' to authenticate using the legacy method")
					os.Exit(1)
				}
				// Use legacy auth
				registryConfig.Username = authConfig.Username
				registryConfig.Password = authConfig.Password
				registryConfig.APIKey = authConfig.APIKey
				registryConfig.AuthType = authConfig.AuthType
			}

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

			// Publish to registry with authentication
			client := registry.NewClient(registryConfig.URL)
			if registryConfig.AuthType == "token" {
				client.SetAPIKey(registryConfig.APIKey)
			} else if registryConfig.AuthType == "basic" {
				client.SetBasicAuth(registryConfig.Username, registryConfig.Password)
			}

			cmd.Printf("Publishing %s@%s to %s...\n", m.Package.Name, m.Package.Version, registryConfig.URL)
			if err := client.Publish(archivePath, metadata); err != nil {
				cmd.PrintErrf("Error publishing package: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("Successfully published %s@%s!\n", m.Package.Name, m.Package.Version)
		},
	}
	root.AddCommand(publishCmd)

	// Publish test command
	publishTestCmd := &cobra.Command{
		Use:   "publish-test",
		Short: "Publish a package to the registry (test mode)",
		Run: func(cmd *cobra.Command, args []string) {
			// Check authentication first
			authService := auth.New(cfg)
			authConfig, err := authService.GetAuthConfig()
			if err != nil {
				cmd.PrintErrf("Authentication required: %v\n", err)
				cmd.PrintErrln("Run 'bifrost login' to authenticate first")
				os.Exit(1)
			}

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

			// Publish to registry with authentication
			client := registry.NewClient(cfg.RegistryURL)
			if authConfig.AuthType == "token" {
				client.SetAPIKey(authConfig.APIKey)
			} else if authConfig.AuthType == "basic" {
				client.SetBasicAuth(authConfig.Username, authConfig.Password)
			}

			cmd.Printf("Publishing %s@%s to %s (test mode)...\n", m.Package.Name, m.Package.Version, cfg.RegistryURL)
			if err := client.PublishTest(archivePath, metadata); err != nil {
				cmd.PrintErrf("Error publishing package: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("Successfully published %s@%s (test mode)!\n", m.Package.Name, m.Package.Version)
		},
	}
	root.AddCommand(publishTestCmd)

	// Config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Bifrost configuration",
		Long:  "Configure registry settings, authentication, and user information",
	}

	// Config set command
	configSetCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value. Available keys:
  registry.url       - Registry URL
  registry.username  - Registry username for basic auth
  registry.password  - Registry password for basic auth
  registry.api-key   - API key for token auth
  registry.auth-type - Authentication type (basic, token, none)
  user.name          - Your name
  user.email         - Your email address`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]

			userConfig, err := cfg.LoadUserConfig()
			if err != nil {
				cmd.PrintErrf("Error loading config: %v\n", err)
				os.Exit(1)
			}

			switch key {
			case "registry.url":
				userConfig.Registry.URL = value
			case "registry.username":
				userConfig.Registry.Username = value
				if userConfig.Registry.AuthType == "" {
					userConfig.Registry.AuthType = "basic"
				}
			case "registry.password":
				userConfig.Registry.Password = value
				if userConfig.Registry.AuthType == "" {
					userConfig.Registry.AuthType = "basic"
				}
			case "registry.api-key":
				userConfig.Registry.APIKey = value
				userConfig.Registry.AuthType = "token"
			case "registry.auth-type":
				if value != "basic" && value != "token" && value != "none" {
					cmd.PrintErrln("Error: auth-type must be 'basic', 'token', or 'none'")
					os.Exit(1)
				}
				userConfig.Registry.AuthType = value
			case "user.name":
				userConfig.User.Name = value
			case "user.email":
				userConfig.User.Email = value
			default:
				cmd.PrintErrf("Error: unknown config key '%s'\n", key)
				os.Exit(1)
			}

			if err := cfg.SaveUserConfig(userConfig); err != nil {
				cmd.PrintErrf("Error saving config: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("Set %s = %s\n", key, value)
		},
	}
	configCmd.AddCommand(configSetCmd)

	// Config get command
	configGetCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get configuration value(s)",
		Long:  "Get a specific configuration value or show all configuration",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			userConfig, err := cfg.LoadUserConfig()
			if err != nil {
				cmd.PrintErrf("Error loading config: %v\n", err)
				os.Exit(1)
			}

			if len(args) == 0 {
				// Show all config
				cmd.Println("Registry configuration:")
				cmd.Printf("  url: %s\n", userConfig.Registry.URL)
				cmd.Printf("  auth-type: %s\n", userConfig.Registry.AuthType)
				if userConfig.Registry.Username != "" {
					cmd.Printf("  username: %s\n", userConfig.Registry.Username)
				}
				if userConfig.Registry.APIKey != "" {
					cmd.Printf("  api-key: %s\n", maskAPIKey(userConfig.Registry.APIKey))
				}
				
				if userConfig.User.Name != "" || userConfig.User.Email != "" {
					cmd.Println("\nUser information:")
					if userConfig.User.Name != "" {
						cmd.Printf("  name: %s\n", userConfig.User.Name)
					}
					if userConfig.User.Email != "" {
						cmd.Printf("  email: %s\n", userConfig.User.Email)
					}
				}
			} else {
				// Show specific key
				key := args[0]
				var value string
				
				switch key {
				case "registry.url":
					value = userConfig.Registry.URL
				case "registry.username":
					value = userConfig.Registry.Username
				case "registry.password":
					value = "***" // Never show password
				case "registry.api-key":
					value = maskAPIKey(userConfig.Registry.APIKey)
				case "registry.auth-type":
					value = userConfig.Registry.AuthType
				case "user.name":
					value = userConfig.User.Name
				case "user.email":
					value = userConfig.User.Email
				default:
					cmd.PrintErrf("Error: unknown config key '%s'\n", key)
					os.Exit(1)
				}
				
				cmd.Printf("%s = %s\n", key, value)
			}
		},
	}
	configCmd.AddCommand(configGetCmd)

	// Config unset command
	configUnsetCmd := &cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a configuration value",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]

			userConfig, err := cfg.LoadUserConfig()
			if err != nil {
				cmd.PrintErrf("Error loading config: %v\n", err)
				os.Exit(1)
			}

			switch key {
			case "registry.username":
				userConfig.Registry.Username = ""
			case "registry.password":
				userConfig.Registry.Password = ""
			case "registry.api-key":
				userConfig.Registry.APIKey = ""
			case "user.name":
				userConfig.User.Name = ""
			case "user.email":
				userConfig.User.Email = ""
			default:
				cmd.PrintErrf("Error: cannot unset '%s' or key does not exist\n", key)
				os.Exit(1)
			}

			if err := cfg.SaveUserConfig(userConfig); err != nil {
				cmd.PrintErrf("Error saving config: %v\n", err)
				os.Exit(1)
			}

			cmd.Printf("Unset %s\n", key)
		},
	}
	configCmd.AddCommand(configUnsetCmd)

	root.AddCommand(configCmd)

	// Version command
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show Bifrost version",
		Run: func(cmd *cobra.Command, args []string) {
			// Try to read version from Bifrost's own manifest
			var versionStr string
			
			// Look for Bifrost.toml in various locations
			possiblePaths := []string{
				"Bifrost.toml",
				filepath.Join(filepath.Dir(os.Args[0]), "Bifrost.toml"),
				filepath.Join(filepath.Dir(os.Args[0]), "..", "Bifrost.toml"),
				"/usr/local/share/bifrost/Bifrost.toml",
				"/usr/share/bifrost/Bifrost.toml",
			}
			
			// If installed via go install, check GOPATH
			if gopath := os.Getenv("GOPATH"); gopath != "" {
				possiblePaths = append(possiblePaths, 
					filepath.Join(gopath, "src", "github.com", "javanhut", "bifrost", "Bifrost.toml"))
			}
			
			// Check each possible location
			for _, path := range possiblePaths {
				if m, err := manifest.Load(path); err == nil {
					versionStr = m.Package.Version
					break
				}
			}
			
			// If no external manifest found, try embedded one
			if versionStr == "" && embeddedManifest != "" {
				// Parse embedded manifest
				tempFile, err := os.CreateTemp("", "bifrost-manifest-*.toml")
				if err == nil {
					defer os.Remove(tempFile.Name())
					defer tempFile.Close()
					
					if _, err := tempFile.WriteString(embeddedManifest); err == nil {
						if m, err := manifest.Load(tempFile.Name()); err == nil {
							versionStr = m.Package.Version
						}
					}
				}
			}
			
			// Ultimate fallback
			if versionStr == "" {
				versionStr = "unknown"
			}
			
			cmd.Printf("Bifrost %s\n", versionStr)
			cmd.Println("Package manager for the Carrion programming language")
			
			registryConfig, err := cfg.GetRegistryConfig()
			if err != nil {
				cmd.Printf("Registry: %s (default)\n", cfg.RegistryURL)
			} else {
				cmd.Printf("Registry: %s\n", registryConfig.URL)
			}
		},
	})

	cobra.CheckErr(root.Execute())
}
