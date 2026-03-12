package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/globaltask/bank/internal/infrastructure/utils"
	"github.com/google/uuid"
)

type SupabaseIdentityService struct {
	url            string
	serviceRoleKey string
}

func NewSupabaseIdentityService() *SupabaseIdentityService {
	return &SupabaseIdentityService{
		url:            os.Getenv("SUPABASE_URL"),
		serviceRoleKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
	}
}

func (s *SupabaseIdentityService) RegisterUser(ctx context.Context, name string, doc string, countryID int) (uuid.UUID, error) {
	// 1. Register user in Supabase Auth (Admin API)
	// Using Service Role Key allows us to use the admin API directly.
	log.Printf("Registering user: %s, doc: %s", utils.MaskName(name), utils.MaskID(doc))

	email := fmt.Sprintf("%s.%s@bank.internal", strings.ToLower(strings.ReplaceAll(name, " ", ".")), doc)
	password := os.Getenv("DEFAULT_USER_PASSWORD")
	if password == "" {
		password = "Password123!" // Fallback if not set
	}

	userData := map[string]interface{}{
		"email":         email,
		"password":      password,
		"email_confirm": true, // Automatically confirm email
		"user_metadata": map[string]interface{}{
			"full_name":         name,
			"identity_document": doc,
			"country_id":        countryID,
			"role":              "USER", // Automatically registered borrowers are always USERS
		},
	}

	jsonData, err := json.Marshal(userData)
	if err != nil {
		return uuid.Nil, err
	}

	// For admin user creation, the endpoint is different
	url := strings.TrimSuffix(s.url, "/")
	authUrl := fmt.Sprintf("%s/auth/v1/admin/users", url)
	req, err := http.NewRequestWithContext(ctx, "POST", authUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return uuid.Nil, err
	}

	req.Header.Set("apikey", s.serviceRoleKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return uuid.Nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Failed to sign up user. Status: %d", resp.StatusCode)
		return uuid.Nil, fmt.Errorf("failed to sign up user: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to read signup response: %v", err)
	}

	var signupResp struct {
		ID uuid.UUID `json:"id"`
	}
	if err := json.Unmarshal(body, &signupResp); err != nil {
		return uuid.Nil, fmt.Errorf("failed to decode signup response: %v", err)
	}

	if signupResp.ID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("failed to retrieve user ID from signup response")
	}
	log.Printf("User registered successfully with ID: %s", signupResp.ID)

	return signupResp.ID, nil
}
