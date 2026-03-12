# SKILL: Bank Provider Scaffolder
==============================

This skill guides the creation of a new bank provider integration in the GlobalTask Bank project. It ensures the integration follows the **Open/Closed Principle** and uses the **Strategy Pattern** via the `ProviderClient` interface.

## 🕵️ Discovery Questions
Before generating code, you MUST ask the user these questions to define the integration details:

1.  **Provider Name**: What is the name of the bank or service? (e.g., 'BBVA', 'Revolut')
2.  **Base URL**: What is the root API address? (e.g., 'https://api.bbva.com/v2')
3.  **Authentication**: How do we authenticate?
    *   API Key / Token in Header (Name: ?, Value: ?)
    *   OAuth2 (Client ID, Client Secret, Token URL)
    *   Basic Auth (Username, Password)
4.  **Endpoint Paths**: What are the specific paths for:
    *   `FETCH_BANK_DATA` (e.g., '/accounts/summary')
    *   `VALIDATE_USER_IDENTITY` (e.g., '/users/verify')
5.  **Response Mapping**: Does the provider return fields that differ from our `entity.JSONB` standard? (e.g., 'user_id' vs 'borrower_id')

---

## 🏗️ Implementation Template

Create a new file at `backend/internal/domain/workflow/{provider_name}_provider_client.go` with the following content:

```go
package workflow

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type {ProviderName}ProviderClient struct {
	client *http.Client
	apiKey string // Add other auth fields as needed
}

func New{ProviderName}ProviderClient(apiKey string) *{ProviderName}ProviderClient {
	return &{ProviderName}ProviderClient{
		client: &http.Client{Timeout: 15 * time.Second},
		apiKey: apiKey,
	}
}

func (p *{ProviderName}ProviderClient) Execute(ctx context.Context, baseURL string, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	// 1. Construct Full URL
	fullURL := baseURL + endpoint

	// 2. Add Auth Logic here (e.g., p.client.Do with headers)
	// 3. Perform the request and return standard JSONB
	return map[string]interface{}{
		"status": "success",
		"provider": "{ProviderName}",
		// Add mapped fields here
	}, nil
}
```

---

## 🔗 Registration Guide
1.  Open `backend/internal/domain/workflow/provider_client.go`.
2.  Update `RegisterDefaultClients` to include the new client:
    ```go
    func RegisterDefaultClients(factory ProviderFactory) {
        // ... existing registrations
        factory.Register("{ProviderName}", New{ProviderName}ProviderClient(os.Getenv("{PROVIDER}_API_KEY")))
    }
    ```

---

## 🗄️ Database Mapping (SQL)
Provide the user with this SQL template to run in Supabase:

```sql
-- 1. Create the Provider
INSERT INTO public.bank_providers (country_id, provider_name, base_url, api_config)
VALUES ({country_id}, '{ProviderName}', '{base_url}', '{}');

-- 2. Map Workflow Steps
INSERT INTO public.workflow_providers (workflow_name, provider_id, event_step, endpoint_path)
VALUES 
    ('{ISO_CODE}', (SELECT id FROM bank_providers WHERE provider_name = '{ProviderName}'), 'FETCH_BANK_DATA', '{fetch_path}'),
    ('{ISO_CODE}', (SELECT id FROM bank_providers WHERE provider_name = '{ProviderName}'), 'VALIDATE_USER_IDENTITY', '{validate_path}');
```

---

## ✅ Verification Steps
1.  Verify the new file compiles.
2.  Add a unit test in `backend/internal/domain/workflow/workflow_test.go` mocking the new client.
3.  Check the logs when the `WorkflowEngine` triggers the new step.
