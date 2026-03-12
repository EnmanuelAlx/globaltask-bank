package middleware

import (
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

type JWTAuthConfig struct {
	JWKSURL   string
	JWTSecret string
}

var (
	jwtConfig JWTAuthConfig
	kf        keyfunc.Keyfunc
	initOnce  sync.Once
)

func SetJWTSecret(secret string) {
	jwtConfig.JWTSecret = secret
}

func InitJWKS() {
	initOnce.Do(func() {
		jwksURL := os.Getenv("SUPABASE_JWKS_URL")
		if jwksURL == "" {
			jwksURL = jwtConfig.JWKSURL
		}
		if jwksURL == "" {
			log.Println("SUPABASE_JWKS_URL not set, falling back to HS256 if secret is provided")
			return
		}

		var err error
		kf, err = keyfunc.NewDefault([]string{jwksURL})
		if err != nil {
			log.Printf("Failed to create keyfunc from JWKS URL: %v", err)
		}
	})
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

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Try RS256 or ES256 with JWKS first
			if _, ok := token.Method.(*jwt.SigningMethodRSA); ok || strings.HasPrefix(token.Method.Alg(), "ES") || strings.HasPrefix(token.Method.Alg(), "RS") {
				if kf == nil {
					InitJWKS()
				}
				if kf != nil {
					key, err := kf.Keyfunc(token)
					if err != nil {
						log.Printf("Keyfunc failed for alg %v: %v", token.Method.Alg(), err)
					}
					return key, err
				}
			}

			// Fallback to HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("Invalid signing method: %v", token.Header["alg"])
				return nil, jwt.ErrSignatureInvalid
			}
			// Use JWT_SECRET or fall back to config
			secret := jwtConfig.JWTSecret
			if secret == "" {
				secret = os.Getenv("JWT_SECRET")
			}
			if secret == "" {
				secret = "default-secret-change-me"
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			log.Printf("JWT validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// Get user ID from claims
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

		// Set userID first so it can be used by GetUserID if needed
		c.Set("user_id", userID)

		// Get role from profile table instead of relying only on claims
		role := "USER"
		if profileRepo != nil {
			profile, err := profileRepo.GetByID(c.Request.Context(), userID)
			if err == nil && profile != nil {
				role = profile.Role
			} else if err != nil {
				log.Printf("Failed to fetch profile for user %s: %v", userID, err)
			}
		} else {
			// Fallback to claims if no profile repository provided
			if roleClaim, ok := claims["role"].(string); ok {
				role = roleClaim
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
