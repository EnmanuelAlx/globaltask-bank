package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/globaltask/bank/internal/infrastructure/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	kf       keyfunc.Keyfunc
	issuer   string
	initOnce sync.Once
	initErr  error
)

// InitJWKS initializes the JWKS keyfunc from SUPABASE_JWKS_URL.
// It must be called once at application startup. Returns an error if
// the required environment variables are missing or JWKS fetch fails.
func InitJWKS() error {
	initOnce.Do(func() {
		jwksURL := os.Getenv("SUPABASE_JWKS_URL")
		if jwksURL == "" {
			initErr = fmt.Errorf("SUPABASE_JWKS_URL environment variable is required")
			return
		}

		supabaseURL := os.Getenv("SUPABASE_URL")
		if supabaseURL == "http://host.docker.internal:54321" {
			supabaseURL = "http://127.0.0.1:54321"
		}
		if supabaseURL == "" {
			initErr = fmt.Errorf("SUPABASE_URL environment variable is required")
			return
		}
		issuer = strings.TrimSuffix(supabaseURL, "/") + "/auth/v1"

		kf, initErr = keyfunc.NewDefault([]string{jwksURL})
		if initErr != nil {
			initErr = fmt.Errorf("failed to fetch JWKS from %s: %w", jwksURL, initErr)
		}
	})
	return initErr
}

// JWTAuth middleware validates JWT tokens from Supabase and fetches the user's role from their profile.
func JWTAuth(profileRepo repository.ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		if kf == nil {
			log.Println("JWKS not initialized, rejecting request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication service unavailable"})
			c.Abort()
			return
		}

		// Parse and validate token with strict options:
		// - Only asymmetric algorithms allowed (RS256, ES256)
		// - Issuer must match our Supabase project
		// - Audience must be "authenticated"
		// - Expiration is validated by default
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify it's an asymmetric algorithm (RSA or ECDSA)
			switch token.Method.(type) {
			case *jwt.SigningMethodRSA, *jwt.SigningMethodECDSA:
				return kf.Keyfunc(token)
			default:
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		},
			jwt.WithIssuer(issuer),
			jwt.WithAudience("authenticated"),
			jwt.WithExpirationRequired(),
			jwt.WithValidMethods([]string{"RS256", "ES256"}),
		)

		if err != nil || !token.Valid {
			log.Printf("JWT validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing subject in token"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)

		role := "USER"
		if profileRepo != nil {
			profile, err := profileRepo.GetByID(c.Request.Context(), userID)
			if err == nil && profile != nil {
				role = profile.Role
			} else if err != nil {
				log.Printf("Failed to fetch profile for user %s: %v", userID, err)
			}
		}

		c.Set("user_role", role)
		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) uuid.UUID {
	if id, exists := c.Get("user_id"); exists {
		return id.(uuid.UUID)
	}
	return uuid.Nil
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("user_role"); exists {
		return role.(string)
	}
	return "USER"
}
