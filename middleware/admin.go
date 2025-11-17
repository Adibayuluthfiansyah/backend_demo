package middleware

import (
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRaw, exists := c.Get("user")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		user := userRaw.(models.User)

		if user.Role != "admin" {
			c.JSON(403, gin.H{
				"error": "Hanya admin yang dapat mengakses endpoint ini",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
