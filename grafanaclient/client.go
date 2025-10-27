package grafanaclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Grafana API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	username   string
	password   string
}

// NewClient creates a new Grafana client
func NewClient(baseURL string, apiKey string) *Client {
	if apiKey == "" {
		// Will use basic auth
		return &Client{
			baseURL: baseURL,
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// NewClientWithAuth creates a new Grafana client with basic auth
func NewClientWithAuth(baseURL string, username string, password string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		username: username,
		password: password,
	}
}

// Datasource represents a Grafana datasource
type Datasource struct {
	ID       int64                  `json:"id"`
	UID      string                 `json:"uid"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	URL      string                 `json:"url"`
	JSONData map[string]interface{} `json:"jsonData,omitempty"`
}

// Dashboard represents a Grafana dashboard
type Dashboard struct {
	UID           string                 `json:"uid,omitempty"`
	ID            int64                  `json:"id,omitempty"`
	Title         string                 `json:"title"`
	Tags          []string               `json:"tags,omitempty"`
	Time          interface{}            `json:"time,omitempty"`
	Refresh       interface{}            `json:"refresh,omitempty"`
	SchemaVersion int                    `json:"schemaVersion,omitempty"`
	Version       int                    `json:"version,omitempty"`
	Dashboard     map[string]interface{} `json:"dashboard,omitempty"`
	Panels        interface{}            `json:"panels,omitempty"`
}

// AlertRule represents a Grafana alert rule
type AlertRule struct {
	UID          string                 `json:"uid,omitempty"`
	Title        string                 `json:"title"`
	Condition    string                 `json:"condition"`
	Data         []interface{}          `json:"data"`
	ExecErrState string                 `json:"execErrState,omitempty"`
	For          string                 `json:"for,omitempty"`
	NoDataState  string                 `json:"noDataState,omitempty"`
	Annotations  map[string]interface{} `json:"annotations,omitempty"`
	Labels       map[string]interface{} `json:"labels,omitempty"`
}

// CreateDatasource creates a new datasource in Grafana
func (c *Client) CreateDatasource(datasource Datasource) (*Datasource, error) {
	url := fmt.Sprintf("%s/api/datasources", c.baseURL)

	data, err := json.Marshal(datasource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal datasource: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create datasource: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result Datasource
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListDatasources lists all datasources in Grafana
func (c *Client) ListDatasources() ([]Datasource, error) {
	url := fmt.Sprintf("%s/api/datasources", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list datasources: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result []Datasource
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetDatasource gets a datasource by ID
func (c *Client) GetDatasource(id int64) (*Datasource, error) {
	url := fmt.Sprintf("%s/api/datasources/%d", c.baseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get datasource: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result Datasource
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CreateDashboard creates a new dashboard in Grafana
func (c *Client) CreateDashboard(dashboard Dashboard) (*Dashboard, error) {
	url := fmt.Sprintf("%s/api/dashboards/db", c.baseURL)

	data, err := json.Marshal(dashboard)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dashboard: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create dashboard: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dashboard, nil
}

// CreateAlertRule creates a new alert rule in Grafana
func (c *Client) CreateAlertRule(rule AlertRule) error {
	url := fmt.Sprintf("%s/api/ruler/grafana/api/v1/rules", c.baseURL)

	data, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("failed to marshal alert rule: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create alert rule: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetHealth checks the health of the Grafana instance
func (c *Client) GetHealth() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/health", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("health check failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// setAuthHeaders sets the appropriate authentication headers
func (c *Client) setAuthHeaders(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	} else if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
}
