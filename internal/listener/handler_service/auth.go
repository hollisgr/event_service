package handler_service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckHeader(header string, token string) error {
	auth := fmt.Sprintf("Bearer %s", token)
	if header != auth {
		return fmt.Errorf("auth error")
	}
	return nil
}

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
