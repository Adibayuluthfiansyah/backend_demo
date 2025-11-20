package controllers

import (
	"net/http"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

// Ambil semua notifikasi user yang login
func GetNotifications(c *gin.Context) {
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User tidak terautentikasi",
		})
		return
	}

	user, ok := userRaw.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Tipe data user di context tidak valid",
		})
		return
	}
	userIDStr := user.ID

	var notifications []models.Notification

	// Urutkan notifikasi dari yang terbaru
	if err := config.DB.Where("user_id = ?", userIDStr).
		Order("created_at DESC").
		Limit(50).
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil notifikasi",
		})
		return
	}
	var unreadCount int64
	config.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userIDStr, false).
		Count(&unreadCount)

	if notifications == nil {
		notifications = []models.Notification{}
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}

// Tandai notifikasi sebagai sudah dibaca
func MarkNotificationAsRead(c *gin.Context) {
	notificationID := c.Param("id")
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User tidak terautentikasi",
		})
		return
	}

	user, ok := userRaw.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Tipe data user di context tidak valid",
		})
		return
	}
	userIDStr := user.ID

	var notification models.Notification

	// Cari notifikasi milik user
	if err := config.DB.Where("id = ? AND user_id = ?", notificationID, userIDStr).
		First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Notifikasi tidak ditemukan",
		})
		return
	}

	// Update jika sudah dibaca
	if !notification.IsRead {
		notification.IsRead = true
		if err := config.DB.Save(&notification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memperbarui notifikasi",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notifikasi berhasil ditandai sebagai dibaca",
	})
}

// Helper function untuk membuat notifikasi
func CreateNotification(userID, message, link string) error {
	notification := models.Notification{
		UserID:  userID,
		Message: message,
		Link:    link,
		IsRead:  false,
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		return err
	}

	return nil
}
