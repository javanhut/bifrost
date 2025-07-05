package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/javanhut/bifrost/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		HomeDir:     "/test/home",
		RegistryURL: "https://test.registry.com",
	}

	service := New(cfg)

	if service == nil {
		t.Fatal("New() returned nil")
	}
	if service.config != cfg {
		t.Errorf("service.config = %v, want %v", service.config, cfg)
	}
}

func TestService_tryTokenAuth(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	tests := []struct {
		name     string
		handler  http.HandlerFunc
		username string
		password string
		wantErr  bool
		checkAuth func(t *testing.T, auth *config.AuthConfig)
	}{
		{
			name: "successful token auth",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/auth/login" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("unexpected method: %s", r.Method)
				}

				var req LoginRequest
				json.NewDecoder(r.Body).Decode(&req)
				
				if req.Username != "testuser" || req.Password != "testpass" {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid credentials"})
					return
				}

				resp := LoginResponse{
					APIKey:   "test-api-key",
					Username: req.Username,
					Message:  "Welcome!",
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			username: "testuser",
			password: "testpass",
			wantErr:  false,
			checkAuth: func(t *testing.T, auth *config.AuthConfig) {
				if auth.APIKey != "test-api-key" {
					t.Errorf("auth.APIKey = %v, want test-api-key", auth.APIKey)
				}
				if auth.AuthType != "token" {
					t.Errorf("auth.AuthType = %v, want token", auth.AuthType)
				}
			},
		},
		{
			name: "no token auth endpoint",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			username: "testuser",
			password: "testpass",
			wantErr:  true,
		},
		{
			name: "unauthorized",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "bad credentials"})
			},
			username: "testuser",
			password: "wrongpass",
			wantErr:  true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			username: "testuser",
			password: "testpass",
			wantErr:  true,
		},
		{
			name: "empty API key response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := LoginResponse{
					Username: "testuser",
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			username: "testuser",
			password: "testpass",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			cfg := &config.Config{
				HomeDir:     tempDir,
				AuthFile:    authFile,
				RegistryURL: server.URL,
			}
			service := New(cfg)

			err := service.tryTokenAuth(tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("tryTokenAuth() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.checkAuth != nil {
				auth, _ := cfg.LoadAuth()
				tt.checkAuth(t, auth)
			}

			// Clean up auth file
			os.Remove(authFile)
		})
	}
}

func TestService_testBasicAuth(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "successful health check",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/health" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				
				// Check basic auth header
				username, password, ok := r.BasicAuth()
				if !ok || username != "testuser" || password != "testpass" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name: "health check fails",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			cfg := &config.Config{
				RegistryURL: server.URL,
			}
			service := New(cfg)

			err := service.testBasicAuth("testuser", "testpass")
			if (err != nil) != tt.wantErr {
				t.Errorf("testBasicAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_saveBasicAuth(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	cfg := &config.Config{
		HomeDir:     tempDir,
		AuthFile:    authFile,
		RegistryURL: "https://test.registry.com",
	}
	service := New(cfg)

	err := service.saveBasicAuth("testuser", "testpass")
	if err != nil {
		t.Fatalf("saveBasicAuth() error = %v", err)
	}

	// Verify auth was saved correctly
	auth, err := cfg.LoadAuth()
	if err != nil {
		t.Fatalf("failed to load auth: %v", err)
	}

	if auth.Username != "testuser" {
		t.Errorf("auth.Username = %v, want testuser", auth.Username)
	}
	if auth.Password != "testpass" {
		t.Errorf("auth.Password = %v, want testpass", auth.Password)
	}
	if auth.AuthType != "basic" {
		t.Errorf("auth.AuthType = %v, want basic", auth.AuthType)
	}
	if auth.Registry != "https://test.registry.com" {
		t.Errorf("auth.Registry = %v, want https://test.registry.com", auth.Registry)
	}
}

func TestService_Logout(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	// Create an auth file
	authData := &config.AuthConfig{
		Username: "testuser",
		APIKey:   "test-key",
		AuthType: "token",
	}
	data, _ := json.Marshal(authData)
	os.WriteFile(authFile, data, 0600)

	cfg := &config.Config{
		HomeDir:  tempDir,
		AuthFile: authFile,
	}
	service := New(cfg)

	err := service.Logout()
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	// Verify auth file was removed
	if _, err := os.Stat(authFile); !os.IsNotExist(err) {
		t.Error("auth file should have been removed")
	}
}

func TestService_IsAuthenticated(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	tests := []struct {
		name      string
		authData  *config.AuthConfig
		registry  string
		wantAuth  bool
		wantErr   bool
	}{
		{
			name: "valid token auth",
			authData: &config.AuthConfig{
				Username: "testuser",
				APIKey:   "test-key",
				Registry: "https://test.registry.com",
				AuthType: "token",
			},
			registry: "https://test.registry.com",
			wantAuth: true,
			wantErr:  false,
		},
		{
			name: "valid basic auth",
			authData: &config.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				Registry: "https://test.registry.com",
				AuthType: "basic",
			},
			registry: "https://test.registry.com",
			wantAuth: true,
			wantErr:  false,
		},
		{
			name:     "no auth file",
			authData: nil,
			registry: "https://test.registry.com",
			wantAuth: false,
			wantErr:  false,
		},
		{
			name: "empty username",
			authData: &config.AuthConfig{
				APIKey:   "test-key",
				Registry: "https://test.registry.com",
				AuthType: "token",
			},
			registry: "https://test.registry.com",
			wantAuth: false,
			wantErr:  false,
		},
		{
			name: "different registry",
			authData: &config.AuthConfig{
				Username: "testuser",
				APIKey:   "test-key",
				Registry: "https://other.registry.com",
				AuthType: "token",
			},
			registry: "https://test.registry.com",
			wantAuth: false,
			wantErr:  true,
		},
		{
			name: "missing API key for token auth",
			authData: &config.AuthConfig{
				Username: "testuser",
				Registry: "https://test.registry.com",
				AuthType: "token",
			},
			registry: "https://test.registry.com",
			wantAuth: false,
			wantErr:  true,
		},
		{
			name: "missing password for basic auth",
			authData: &config.AuthConfig{
				Username: "testuser",
				Registry: "https://test.registry.com",
				AuthType: "basic",
			},
			registry: "https://test.registry.com",
			wantAuth: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing auth file
			os.Remove(authFile)

			if tt.authData != nil {
				data, _ := json.Marshal(tt.authData)
				os.WriteFile(authFile, data, 0600)
			}

			cfg := &config.Config{
				HomeDir:     tempDir,
				AuthFile:    authFile,
				RegistryURL: tt.registry,
			}
			service := New(cfg)

			gotAuth, gotAuthData, err := service.IsAuthenticated()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAuthenticated() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotAuth != tt.wantAuth {
				t.Errorf("IsAuthenticated() gotAuth = %v, want %v", gotAuth, tt.wantAuth)
			}
			if tt.authData != nil && gotAuthData == nil && tt.wantAuth {
				t.Error("expected auth data but got nil")
			}
		})
	}
}

func TestService_GetAuthConfig(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	tests := []struct {
		name     string
		authData *config.AuthConfig
		wantErr  bool
	}{
		{
			name: "authenticated user",
			authData: &config.AuthConfig{
				Username: "testuser",
				APIKey:   "test-key",
				Registry: "https://test.registry.com",
				AuthType: "token",
			},
			wantErr: false,
		},
		{
			name:     "not authenticated",
			authData: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(authFile)

			if tt.authData != nil {
				data, _ := json.Marshal(tt.authData)
				os.WriteFile(authFile, data, 0600)
			}

			cfg := &config.Config{
				HomeDir:     tempDir,
				AuthFile:    authFile,
				RegistryURL: "https://test.registry.com",
			}
			service := New(cfg)

			got, err := service.GetAuthConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got == nil {
				t.Error("expected auth config but got nil")
			}
		})
	}
}

func TestService_GetAPIKey(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	tests := []struct {
		name     string
		authData *config.AuthConfig
		wantKey  string
		wantErr  bool
	}{
		{
			name: "token auth with API key",
			authData: &config.AuthConfig{
				Username: "testuser",
				APIKey:   "test-api-key",
				Registry: "https://test.registry.com",
				AuthType: "token",
			},
			wantKey: "test-api-key",
			wantErr: false,
		},
		{
			name: "basic auth",
			authData: &config.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				Registry: "https://test.registry.com",
				AuthType: "basic",
			},
			wantKey: "",
			wantErr: true,
		},
		{
			name:     "not authenticated",
			authData: nil,
			wantKey:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(authFile)

			if tt.authData != nil {
				data, _ := json.Marshal(tt.authData)
				os.WriteFile(authFile, data, 0600)
			}

			cfg := &config.Config{
				HomeDir:     tempDir,
				AuthFile:    authFile,
				RegistryURL: "https://test.registry.com",
			}
			service := New(cfg)

			got, err := service.GetAPIKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantKey {
				t.Errorf("GetAPIKey() = %v, want %v", got, tt.wantKey)
			}
		})
	}
}

func TestService_Login_Integration(t *testing.T) {
	// This test demonstrates the Login method's behavior with mocked HTTP responses
	// We cannot fully test the interactive input portions due to terminal dependencies
	
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "successful token auth scenario",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth/login" {
					// Simulate successful token auth
					resp := LoginResponse{
						APIKey:   "test-token",
						Username: "testuser",
						Message:  "Login successful",
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(resp)
				} else if r.URL.Path == "/api/health" {
					w.WriteHeader(http.StatusOK)
				}
			},
		},
		{
			name: "fallback to basic auth scenario", 
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth/login" {
					// No token auth endpoint
					w.WriteHeader(http.StatusNotFound)
				} else if r.URL.Path == "/api/health" {
					// Basic auth health check passes
					w.WriteHeader(http.StatusOK)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			cfg := &config.Config{
				HomeDir:     tempDir,
				AuthFile:    authFile,
				RegistryURL: server.URL,
			}
			service := New(cfg)

			// Note: We cannot actually call Login() due to its interactive nature
			// Instead, we test the underlying authentication methods it uses
			
			// Test the token auth path
			err := service.tryTokenAuth("testuser", "testpass")
			if tt.name == "successful token auth scenario" && err != nil {
				t.Errorf("tryTokenAuth() should succeed for token auth scenario: %v", err)
			}
			
			// Test the basic auth path
			if tt.name == "fallback to basic auth scenario" {
				err := service.testBasicAuth("testuser", "testpass")
				if err != nil {
					t.Errorf("testBasicAuth() should succeed: %v", err)
				}
				
				err = service.saveBasicAuth("testuser", "testpass")
				if err != nil {
					t.Errorf("saveBasicAuth() should succeed: %v", err)
				}
			}

			// Clean up
			os.Remove(authFile)
		})
	}
}

func TestService_ValidateAuth(t *testing.T) {
	tempDir := t.TempDir()
	authFile := filepath.Join(tempDir, "auth.json")

	tests := []struct {
		name     string
		authData *config.AuthConfig
		handler  http.HandlerFunc
		wantErr  bool
		checkAuth func(t *testing.T, authFile string)
	}{
		{
			name: "valid token auth",
			authData: &config.AuthConfig{
				Username: "testuser",
				APIKey:   "valid-key",
				Registry: "will-be-replaced",
				AuthType: "token",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/auth/validate" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				
				auth := r.Header.Get("Authorization")
				if auth != "Bearer valid-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name: "expired token",
			authData: &config.AuthConfig{
				Username: "testuser",
				APIKey:   "expired-key",
				Registry: "will-be-replaced",
				AuthType: "token",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErr: true,
			checkAuth: func(t *testing.T, authFile string) {
				// Auth should be cleared
				if _, err := os.Stat(authFile); !os.IsNotExist(err) {
					t.Error("auth file should have been removed for expired token")
				}
			},
		},
		{
			name: "no validation endpoint",
			authData: &config.AuthConfig{
				Username: "testuser",
				APIKey:   "test-key",
				Registry: "will-be-replaced",
				AuthType: "token",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: false,
		},
		{
			name: "basic auth always valid",
			authData: &config.AuthConfig{
				Username: "testuser",
				Password: "testpass",
				Registry: "will-be-replaced",
				AuthType: "basic",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Should not be called for basic auth
				t.Error("validation endpoint should not be called for basic auth")
			},
			wantErr: false,
		},
		{
			name:     "not authenticated",
			authData: nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Should not be called
				t.Error("validation endpoint should not be called when not authenticated")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(authFile)

			var server *httptest.Server
			if tt.handler != nil {
				server = httptest.NewServer(tt.handler)
				defer server.Close()
			}

			if tt.authData != nil {
				if server != nil {
					tt.authData.Registry = server.URL
				}
				data, _ := json.Marshal(tt.authData)
				os.WriteFile(authFile, data, 0600)
			}

			registryURL := "https://test.registry.com"
			if server != nil {
				registryURL = server.URL
			}

			cfg := &config.Config{
				HomeDir:     tempDir,
				AuthFile:    authFile,
				RegistryURL: registryURL,
			}
			service := New(cfg)

			err := service.ValidateAuth()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuth() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkAuth != nil {
				tt.checkAuth(t, authFile)
			}
		})
	}
}