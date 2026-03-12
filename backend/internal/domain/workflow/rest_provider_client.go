package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RestProviderClient is a real implementation of ProviderClient using HTTP
type RestProviderClient struct {
	client *http.Client
}

func NewRestProviderClient() *RestProviderClient {
	return &RestProviderClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Execute performs a POST request to the combined URL
func (m *RestProviderClient) Execute(ctx context.Context, baseURL string, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	fullURL := baseURL + endpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("provider returned status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Some endpoints might return empty body on success (like 202 Accepted)
		if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNoContent {
			return map[string]interface{}{"status": "accepted"}, nil
		}
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
