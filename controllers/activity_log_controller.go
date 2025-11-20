package controllers

import (
	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"
	"fmt"
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

// HELPER FUNCTION
func LogActivity(userID, userName, action, message string) {
	go func() {
		logEntry := models.ActivityLog{
			UserID:   userID,
			UserName: userName,
			Action:   action,
			Message:  message,
		}
		if err := config.DB.Create(&logEntry).Error; err != nil {
			fmt.Printf("Gagal mencatat activity log: %v\n", err)
		}
	}()
}
