package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"

	"github.com/mocbotau/api-join-sound/internal/database"
	"github.com/mocbotau/api-join-sound/internal/models"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope string `json:"scope"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() gin.HandlerFunc {
	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return adapter.Wrap(middleware.CheckJWT)
}

// ExtractUserIDFromJWT extracts the Discord user ID from the JWT token's subject field
func extractUserIDFromJWT(c *gin.Context) (int64, error) {
	token := c.Request.Context().Value(jwtmiddleware.ContextKey{})
	if token == nil {
		return 0, fmt.Errorf("no token found in context")
	}

	validatedClaims := token.(*validator.ValidatedClaims)
	subject := validatedClaims.RegisteredClaims.Subject

	parts := strings.Split(subject, "|")
	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid token subject format")
	}

	discordUserID := parts[len(parts)-1]
	userID, err := strconv.ParseInt(discordUserID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return userID, nil
}

// EnsureUserAuthorization checks that the JWT user matches the requested user ID
func EnsureUserAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtUserID, err := extractUserIDFromJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to extract user ID from token"})
			c.Abort()
			return
		}

		if userIDParam := c.Param("userId"); userIDParam != "" {
			requestedUserID, err := strconv.ParseInt(userIDParam, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID parameter"})
				c.Abort()
				return
			}

			if jwtUserID != requestedUserID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own resources"})
				c.Abort()
				return
			}
		}

		c.Set("userID", jwtUserID)
		c.Next()
	}
}

// EnsureResourceOwnership checks ownership for resource-based routes (e.g., sound ID)
func EnsureResourceOwnership(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtUserID, err := extractUserIDFromJWT(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to extract user ID from token"})
			c.Abort()
			return
		}

		if soundID := c.Param("soundId"); soundID != "" {
			sound, err := db.GetSoundByID(soundID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Sound not found"})
				c.Abort()
				return
			}

			var user *models.User

			if user, err = db.GetUserByUserGuildID(sound.UserGuildID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
				c.Abort()
				return
			}

			if jwtUserID != user.UserID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You can only access your own resources"})
				c.Abort()
				return
			}
		}

		c.Set("userID", jwtUserID)
		c.Next()
	}
}
