package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	baseURL    string
	apiURL     string
	httpClient *http.Client
}

type PackageInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Authors     []string `json:"authors"`
	License     string   `json:"license"`
	Homepage    string   `json:"homepage"`
	Repository  string   `json:"repository"`
	Keywords    []string `json:"keywords"`
}

type SearchResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Downloads   int    `json:"downloads"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

func NewClient(baseURL string) *Client {
	// Parse the base URL to extract just the host and scheme
	u, err := url.Parse(baseURL)
	if err != nil {
		// If parsing fails, use baseURL as-is
		return &Client{
			baseURL: baseURL,
			apiURL:  baseURL,
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	}
	
	// Always use the root domain for API calls
	apiURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	
	return &Client{
		baseURL: baseURL,
		apiURL:  apiURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Search(query string) ([]SearchResult, error) {
	u, err := url.Parse(c.apiURL + "/api/search")
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
	}

	var results []SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return results, nil
}

func (c *Client) GetPackageInfo(name, version string) (*PackageInfo, error) {
	url := fmt.Sprintf("%s/api/package/%s/%s", c.apiURL, name, version)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get package info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("package %s@%s not found", name, version)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed with status %d: %s", resp.StatusCode, string(body))
	}

	var info PackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode package info: %w", err)
	}

	return &info, nil
}

func (c *Client) GetPackageLatest(name string) (*PackageInfo, error) {
	return c.GetPackageInfo(name, "latest")
}

func (c *Client) Health() error {
	url := c.apiURL + "/api/health"

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("registry unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry health check failed with status %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return fmt.Errorf("failed to decode health response: %w", err)
	}

	if health.Status != "healthy" {
		return fmt.Errorf("registry status: %s", health.Status)
	}

	return nil
}

func (c *Client) Publish(packagePath string, metadata *PackageInfo) error {
	// Use the new enhanced publish endpoint that handles both metadata and file
	return c.publishPackage(packagePath, metadata)
}

func (c *Client) publishPackage(packagePath string, metadata *PackageInfo) error {
	// Open the package file
	file, err := os.Open(packagePath)
	if err != nil {
		return fmt.Errorf("failed to open package file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add package file
	part, err := writer.CreateFormFile("package", filepath.Base(packagePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Add metadata as JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := writer.WriteField("metadata", string(metadataJSON)); err != nil {
		return fmt.Errorf("failed to write metadata field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/api/publish", c.apiURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to publish package: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("publish failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) DownloadPackage(name, version string) (io.ReadCloser, error) {
	// Use the packages download path according to nginx config
	filename := fmt.Sprintf("%s-%s.tar.gz", name, version)
	url := fmt.Sprintf("%s/packages/%s/%s/%s", c.apiURL, name, version, filename)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download package: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, fmt.Errorf("package %s@%s not found", name, version)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}
