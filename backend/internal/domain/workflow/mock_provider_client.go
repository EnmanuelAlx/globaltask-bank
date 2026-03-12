package workflow

import (
	"context"
	"fmt"
)

// MockProviderClient is a mock implementation of ProviderClient for testing
type MockProviderClient struct{}

// Execute logs the request and returns a mock success response
func (m *MockProviderClient) Execute(ctx context.Context, baseURL string, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("[MockProviderClient] Executing request to: %s%s with payload: %v\n", baseURL, endpoint, payload)

	return map[string]interface{}{
		"status": "mock_success",
	}, nil
}
