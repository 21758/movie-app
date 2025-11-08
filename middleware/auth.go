package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}

		if strings.Contains(c.Request.URL.Path, "/ratings") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] != authToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Invalid authorization token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RaterIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" && strings.Contains(c.Request.URL.Path, "/ratings") {
			raterID := c.GetHeader("X-Rater-Id")
			if raterID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing X-Rater-Id header",
				})
				c.Abort()
				return
			}

			c.Set("rater_id", raterID)
		}

		c.Next()
	}
}
