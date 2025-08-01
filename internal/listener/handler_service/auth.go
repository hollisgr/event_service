package handler_service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckHeader validates the Authorization header against the expected bearer token.
// It formats the expected authorization string ("Bearer {token}") and compares it with the received header.
// Returns an error if the headers don't match, otherwise returns nil.
func CheckHeader(header string, token string) error {
	auth := fmt.Sprintf("Bearer %s", token)
	if header != auth {
		return fmt.Errorf("auth error")
	}
	return nil
}

// AuthMiddleware implements a Gin middleware for basic token-based authentication.
// Extracts the Authorization header from the request context and passes it to CheckHeader for validation.
// If validation fails, aborts the request chain with an Unauthorized response; otherwise, proceeds to next handlers.
func AuthMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		err := CheckHeader(authHeader, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "denied", "message": "access token is incorrect"})
			return
		}
		c.Next()
	}
}
