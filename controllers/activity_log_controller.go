package controllers

import (
	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateActivityLog(userID string, userName, action, message string) {
	log := models.ActivityLog{
		UserID:   userID,
		UserName: userName,
		Action:   action,
		Message:  message,
	}
	go func() {
		config.DB.Create(&log)
	}()
}

// API: Mengambil semua data log untuk ditampilkan di dashboard
func GetAllActivityLogs(c *gin.Context) {
	var logs []models.ActivityLog

	// Urutkan dari yang terbaru (DESC)
	if err := config.DB.Order("created_at desc").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil log aktivitas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
	})
}
