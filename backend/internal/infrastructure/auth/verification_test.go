package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/globaltask/bank/internal/infrastructure/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupabaseIdentityService_RegisterUser_PIISafe(t *testing.T) {
	encryptor, err := security.NewAESEncryptor()
	require.NoError(t, err)

	// Mock Supabase Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/auth/v1/admin/users", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		require.NoError(t, err)

		userMetadata := payload["user_metadata"].(map[string]interface{})

		// Verify identity_document is NOT the plain text
		plaintextDoc := "123456789"
		assert.NotEqual(t, plaintextDoc, userMetadata["identity_document"])

		// Verify identity_document can be decrypted
		decrypted, err := encryptor.Decrypt(userMetadata["identity_document"].(string))
		assert.NoError(t, err)
		assert.Equal(t, plaintextDoc, decrypted)

		// Verify identity_document_bidx is present and matches
		expectedBidx := encryptor.GenerateBlindIndex(plaintextDoc)
		assert.Equal(t, expectedBidx, userMetadata["identity_document_bidx"])

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "550e8400-e29b-41d4-a716-446655440000"}`))
	}))
	defer server.Close()

	os.Setenv("SUPABASE_URL", server.URL)
	os.Setenv("SUPABASE_SERVICE_ROLE_KEY", "test-key")
	defer os.Unsetenv("SUPABASE_URL")
	defer os.Unsetenv("SUPABASE_SERVICE_ROLE_KEY")

	service := NewSupabaseIdentityService(encryptor)
	uid, err := service.RegisterUser(context.Background(), "John Doe", "123456789", 1)

	assert.NoError(t, err)
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", uid.String())
}
