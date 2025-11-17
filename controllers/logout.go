package controllers

import (
	"net/http"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

// Struktur request logout
type LogoutRequest struct {
	TokenID string `json:"token_id" binding:"required"`
}

// Logout menghapus token dari tabel secret_tokens

func Logout(c *gin.Context) {
	// Ambil user dari middleware
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}

	user := userRaw.(models.User)
	db := config.DB

	// Hapus semua token user ini
	if err := db.Where("user_id = ?", user.ID).Delete(&models.SecretToken{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout berhasil",
	})
}
