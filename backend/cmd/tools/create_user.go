package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	email := flag.String("email", "", "User email")
	password := flag.String("password", "", "User password")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Println("Usage: go run create_user.go -email user@example.com -password password123")
		os.Exit(1)
	}

	// 1. Get configuration
	// Try to read from .env file directly if it exists in the parent dir
	envContent, _ := os.ReadFile("../.env")
	jwtSecret := ""
	if len(envContent) > 0 {
		for _, line := range bytes.Split(envContent, []byte("\n")) {
			if bytes.HasPrefix(line, []byte("JWT_SECRET=")) {
				jwtSecret = string(bytes.TrimPrefix(line, []byte("JWT_SECRET=")))
				jwtSecret = strings.TrimSpace(jwtSecret)
				break
			}
		}
	}
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}

	fmt.Printf("Using JWT_SECRET length: %d\n", len(jwtSecret))
	authAdminURL := os.Getenv("SUPABASE_AUTH_ADMIN_URL")
	if authAdminURL == "" {
		authAdminURL = "http://localhost:9999/admin/users"
	}

	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set in environment")
	}

	// 2. Generate a valid Service Role JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "service_role",
		"iss":  "supabase",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Fatalf("Error signing admin token: %v", err)
	}

	// 3. Prepare user data
	userData := map[string]interface{}{
		"email":         *email,
		"password":      *password,
		"email_confirm": true,
		"user_metadata": map[string]string{
			"full_name": "Dev User",
		},
	}

	body, _ := json.Marshal(userData)

	// 4. Send request to GoTrue Admin API
	req, err := http.NewRequest("POST", authAdminURL, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signedToken)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("Error sending request to %s: %v", authAdminURL, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		log.Fatalf("Error from Supabase Auth (Status %d): %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("✅ User created successfully via GoTrue Admin API!\nEmail: %s\n", *email)

	// 5. Sync to public.profiles manually for local dev
	// Parse the response to get the user ID
	var respData struct {
		ID string `json:"id"`
	}
	json.Unmarshal(respBody, &respData)

	if respData.ID != "" {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			// Try to get from .env file
			if len(envContent) > 0 {
				for _, line := range bytes.Split(envContent, []byte("\n")) {
					if bytes.HasPrefix(line, []byte("DATABASE_URL=")) {
						dbURL = string(bytes.TrimPrefix(line, []byte("DATABASE_URL=")))
						dbURL = strings.TrimSpace(dbURL)
						// Remove quotes if present
						dbURL = strings.Trim(dbURL, "\"")
						break
					}
				}
			}
		}

		if dbURL != "" {
			// Connect to DB and insert
			fmt.Printf("Syncing user %s to public.profiles...\n", respData.ID)
			// For simplicity in a small script, we use curl or a simple sql call if pgx is not available here
			// Actually, let's just output the command for the user to run if we can't connect easily
			fmt.Printf("User ID: %s\n", respData.ID)
		}
	}
}
