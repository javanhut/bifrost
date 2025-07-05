package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/javanhut/bifrost/internal/config"
	"golang.org/x/term"
)

type Service struct {
	config *config.Config
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	APIKey   string `json:"api_key"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func New(cfg *config.Config) *Service {
	return &Service{config: cfg}
}

// Login authenticates the user with the registry and stores credentials
func (s *Service) Login() error {
	fmt.Printf("Logging in to %s\n\n", s.config.RegistryURL)

	// Get username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Get password securely
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)
	fmt.Println() // Add newline after password input

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Try token-based authentication first (custom endpoint)
	if err := s.tryTokenAuth(username, password); err == nil {
		return nil
	}

	// Fall back to basic authentication (for Nexus-based registries)
	// Test basic auth by trying to publish a test package
	if err := s.testBasicAuth(username, password); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return s.saveBasicAuth(username, password)
}

// tryTokenAuth attempts to authenticate using token-based authentication
func (s *Service) tryTokenAuth(username, password string) error {
	// Prepare login request
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Send login request to registry
	loginURL := s.config.RegistryURL + "/api/auth/login"
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No custom auth endpoint, registry likely uses basic auth
		return fmt.Errorf("no token auth endpoint available")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("authentication failed: %s", errResp.Error)
		}
		return fmt.Errorf("authentication failed: invalid username or password")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("login failed: %s", errResp.Error)
		}
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if loginResp.APIKey == "" {
		return fmt.Errorf("received empty API key from registry")
	}

	// Save token-based authentication credentials
	auth := &config.AuthConfig{
		APIKey:   loginResp.APIKey,
		Username: loginResp.Username,
		Registry: s.config.RegistryURL,
		AuthType: "token",
	}

	if err := s.config.SaveAuth(auth); err != nil {
		return fmt.Errorf("failed to save authentication: %w", err)
	}

	fmt.Printf("Successfully authenticated as %s using token authentication\n", loginResp.Username)
	if loginResp.Message != "" {
		fmt.Printf("%s\n", loginResp.Message)
	}

	return nil
}

// testBasicAuth tests basic authentication by trying to access the health endpoint
func (s *Service) testBasicAuth(username, password string) error {
	// Test basic auth by trying to access the health endpoint
	// This is a simple check to ensure the registry is accessible
	healthURL := s.config.RegistryURL + "/api/health"
	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Add basic auth header
	req.SetBasicAuth(username, password)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	defer resp.Body.Close()

	// Health endpoint should be accessible regardless of auth
	// This just verifies the registry is reachable
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// saveBasicAuth saves basic authentication credentials
func (s *Service) saveBasicAuth(username, password string) error {
	auth := &config.AuthConfig{
		Username: username,
		Password: password,
		Registry: s.config.RegistryURL,
		AuthType: "basic",
	}

	if err := s.config.SaveAuth(auth); err != nil {
		return fmt.Errorf("failed to save authentication: %w", err)
	}

	fmt.Printf("Successfully authenticated as %s using basic authentication\n", username)
	fmt.Println("Credentials saved for publishing packages")

	return nil
}

// Logout removes stored authentication credentials
func (s *Service) Logout() error {
	if err := s.config.ClearAuth(); err != nil {
		return fmt.Errorf("failed to clear authentication: %w", err)
	}

	fmt.Println("Successfully logged out")
	return nil
}

// IsAuthenticated checks if the user is currently authenticated
func (s *Service) IsAuthenticated() (bool, *config.AuthConfig, error) {
	auth, err := s.config.LoadAuth()
	if err != nil {
		return false, nil, fmt.Errorf("failed to load authentication: %w", err)
	}

	if auth == nil || auth.Username == "" {
		return false, nil, nil
	}

	// Check if the stored registry matches current config
	if auth.Registry != s.config.RegistryURL {
		return false, auth, fmt.Errorf("authentication is for different registry (%s vs %s)", auth.Registry, s.config.RegistryURL)
	}

	// Validate we have the right credentials for the auth type
	if auth.AuthType == "token" && auth.APIKey == "" {
		return false, auth, fmt.Errorf("missing API key for token authentication")
	}
	if auth.AuthType == "basic" && auth.Password == "" {
		return false, auth, fmt.Errorf("missing password for basic authentication")
	}

	return true, auth, nil
}

// GetAuthConfig returns the current authentication config if authenticated
func (s *Service) GetAuthConfig() (*config.AuthConfig, error) {
	authenticated, auth, err := s.IsAuthenticated()
	if err != nil {
		return nil, err
	}

	if !authenticated {
		return nil, fmt.Errorf("not authenticated - run 'bifrost login' first")
	}

	return auth, nil
}

// GetAPIKey returns the current API key if using token authentication
func (s *Service) GetAPIKey() (string, error) {
	auth, err := s.GetAuthConfig()
	if err != nil {
		return "", err
	}

	if auth.AuthType != "token" {
		return "", fmt.Errorf("not using token authentication")
	}

	return auth.APIKey, nil
}

// ValidateAuth checks if the stored authentication is still valid with the registry
func (s *Service) ValidateAuth() error {
	auth, err := s.GetAuthConfig()
	if err != nil {
		return err
	}

	// For basic auth, we can't validate without making an actual request
	// The validation will happen during the first publish attempt
	if auth.AuthType == "basic" {
		return nil
	}

	// For token auth, try to validate with the registry
	if auth.AuthType == "token" {
		// Make a test request to validate the API key
		validateURL := s.config.RegistryURL + "/api/auth/validate"
		req, err := http.NewRequest("GET", validateURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create validation request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+auth.APIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to validate authentication: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			// Clear invalid authentication
			s.config.ClearAuth()
			return fmt.Errorf("authentication expired - please run 'bifrost login' again")
		}

		if resp.StatusCode == http.StatusNotFound {
			// No validation endpoint available, assume token is valid
			return nil
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("authentication validation failed with status: %d", resp.StatusCode)
		}
	}

	return nil
}