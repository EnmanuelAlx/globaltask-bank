package workflow

import (
	"context"
	"fmt"
	"sync"
)

// ProviderClient abstracts third-party communication
type ProviderClient interface {
	Execute(ctx context.Context, baseURL string, endpoint string, payload map[string]interface{}) (map[string]interface{}, error)
}

// ProviderFactory manages ProviderClient implementations
type ProviderFactory interface {
	GetClient(providerName string) (ProviderClient, error)
	Register(providerName string, client ProviderClient)
}

type providerFactory struct {
	clients map[string]ProviderClient
	mu      sync.RWMutex
}

func NewProviderFactory() ProviderFactory {
	return &providerFactory{
		clients: make(map[string]ProviderClient),
	}
}

func (f *providerFactory) GetClient(providerName string) (ProviderClient, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	client, ok := f.clients[providerName]
	if !ok {
		return nil, fmt.Errorf("provider client not found: %s", providerName)
	}
	return client, nil
}

func (f *providerFactory) Register(providerName string, client ProviderClient) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.clients[providerName] = client
}

// RegisterDefaultClients registers initial implementations into the factory
func RegisterDefaultClients(factory ProviderFactory) {
	// Register Mock provider for local testing and missing implementations
	mock := &MockProviderClient{}
	rest := NewRestProviderClient()

	factory.Register("MOCK", mock)
	factory.Register("Santander Totta", rest)
	factory.Register("Millennium BCP", rest)
	factory.Register("Bancolombia", rest)
}
