package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func BasicTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token != os.Getenv("BASIC_TOKEN") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Next()
	}
}
