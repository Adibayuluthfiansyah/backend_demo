package middleware

import (
	"net/http"
	"os"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak ditemukan"})
			c.Abort()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "default_secret"
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid"})
			c.Abort()
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		userID := claims["user_id"]

		var user models.User
		if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User tidak valid"})
			c.Abort()
			return
		}

		c.Set("user", user)

		c.Next()
	}
}
