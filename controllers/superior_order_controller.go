package controllers

import (
	"net/http"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

// ======================================================
// CREATE SuperiorOrder
// ======================================================
func CreateSuperiorOrder(c *gin.Context) {
	var input struct {
		DocumentID string   `json:"document_id" binding:"required"`
		UserIDs    []string `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var created []models.SuperiorOrder
	for _, userID := range input.UserIDs {
		order := models.SuperiorOrder{
			DocumentID: input.DocumentID,
			UserID:     userID,
		}
		if err := config.DB.Create(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create record: " + err.Error()})
			return
		}
		created = append(created, order)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "SuperiorOrder created", "data": created})
}

// ======================================================
// GET ALL SuperiorOrders (grouped by document_id)
// ======================================================
func GetSuperiorOrders(c *gin.Context) {
	var orders []models.SuperiorOrder
	if err := config.DB.Preload("User").Preload("Document").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records: " + err.Error()})
		return
	}

	// Map untuk menampung hasil akhir
	type UserInfo struct {
		Name string `json:"name"`
	}

	type DocumentInfo struct {
		DocumentID string     `json:"document_id"`
		Sender     string     `json:"sender"`
		Subject    string     `json:"subject"`
		Users      []UserInfo `json:"users"`
	}

	grouped := make(map[string]*DocumentInfo)

	for _, o := range orders {
		if _, exists := grouped[o.DocumentID]; !exists {
			grouped[o.DocumentID] = &DocumentInfo{
				DocumentID: o.DocumentID,
				Sender:     o.Document.Sender,
				Subject:    o.Document.Subject,
				Users:      []UserInfo{},
			}
		}
		grouped[o.DocumentID].Users = append(grouped[o.DocumentID].Users, UserInfo{Name: o.User.Name})
	}

	// Ubah map menjadi slice untuk response
	var result []DocumentInfo
	for _, v := range grouped {
		result = append(result, *v)
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ======================================================
// GET SuperiorOrders by document_id
// ======================================================
func GetSuperiorOrdersByDocument(c *gin.Context) {
	documentID := c.Param("document_id")
	var orders []models.SuperiorOrder
	if err := config.DB.Preload("User").Preload("Document").Where("document_id = ?", documentID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records: " + err.Error()})
		return
	}

	if len(orders) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No records found for this document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"document_id": documentID, "user_ids": orders})
}

// ======================================================
// UPDATE SuperiorOrder by document_id
// ======================================================
func UpdateSuperiorOrder(c *gin.Context) {
	documentID := c.Param("document_id")

	var input struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Hapus semua user yang terkait dengan document_id
	if err := config.DB.Where("document_id = ?", documentID).Delete(&models.SuperiorOrder{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old records: " + err.Error()})
		return
	}

	// Tambahkan user_id baru
	var created []models.SuperiorOrder
	for _, userID := range input.UserIDs {
		order := models.SuperiorOrder{
			DocumentID: documentID,
			UserID:     userID,
		}
		if err := config.DB.Create(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create record: " + err.Error()})
			return
		}
		created = append(created, order)
	}

	c.JSON(http.StatusOK, gin.H{"message": "SuperiorOrder updated", "data": created})
}

// ======================================================
// DELETE SuperiorOrder by document_id
// ======================================================
func DeleteSuperiorOrder(c *gin.Context) {
	documentID := c.Param("document_id")

	if err := config.DB.Where("document_id = ?", documentID).Delete(&models.SuperiorOrder{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete records: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All SuperiorOrders for document deleted", "document_id": documentID})
}
