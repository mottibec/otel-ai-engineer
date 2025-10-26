package otelclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OtelClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOtelClient(baseURL string, httpClient *http.Client) *OtelClient {
	return &OtelClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

type OtelAgent struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	LastSeen    string            `json:"lastSeen,omitempty"`
	GroupID     string            `json:"groupID,omitempty"`
	GroupName   string            `json:"groupName,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Description string            `json:"description,omitempty"`
}

// AgentListResponse represents the response from the agents API
type AgentListResponse struct {
	Agents      map[string]OtelAgent `json:"agents"`
	TotalCount  int                  `json:"totalCount"`
	ActiveCount int                  `json:"activeCount"`
}

type AgentConfig struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Version int    `json:"version"`
}

// ListAgents returns all connected OTel agents
func (c *OtelClient) ListAgents() ([]OtelAgent, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/agents")
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response AgentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert map to slice
	agents := make([]OtelAgent, 0, len(response.Agents))
	for _, agent := range response.Agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

// GetAgentConfig returns the current configuration for an agent
func (c *OtelClient) GetAgentConfig(agentID string) (*AgentConfig, error) {
	url := fmt.Sprintf("%s/api/v1/agents/%s/config", c.baseURL, agentID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var config AgentConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &config, nil
}

// UpdateAgentConfig sends a new configuration to an agent
func (c *OtelClient) UpdateAgentConfig(agentID string, yamlConfig string) error {
	payload := map[string]string{
		"config": yamlConfig,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/agents/%s/config", c.baseURL, agentID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
