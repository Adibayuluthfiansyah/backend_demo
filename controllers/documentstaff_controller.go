package controllers

import (
	"net/http"
	"path/filepath"
	"strings"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

// ======================================================
// CREATE DOCUMENT STAFF
// ======================================================
func CreateDocumentStaff(c *gin.Context) {
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User tidak terautentikasi"})
		return
	}
	user := userRaw.(models.User)

	subject := c.PostForm("subject")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tidak dapat membuka file"})
		return
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	var resourceType string
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		resourceType = "image"
	case ".pdf":
		resourceType = "raw"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung. Hanya PDF atau gambar"})
		return
	}

	url, publicID, err := config.UploadToCloudinary(src, resourceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload gagal: " + err.Error()})
		return
	}

	document := models.DocumentStaff{
		UserID:       user.ID,
		FileName:     url,
		PublicID:     publicID,
		ResourceType: resourceType,
		Subject:      subject,
	}

	if err := config.DB.Create(&document).Error; err != nil {
		config.DeleteFromCloudinary(publicID, resourceType)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"document_staff": document})
}

// ======================================================
// GET ALL DOCUMENT STAFF
// ======================================================
func GetDocumentStaffs(c *gin.Context) {
	var documents []models.DocumentStaff
	config.DB.Preload("User").Find(&documents)
	c.JSON(http.StatusOK, gin.H{"document_staffs": documents})
}

// ======================================================
// GET DOCUMENT STAFF BY ID
// ======================================================
func GetDocumentStaffByID(c *gin.Context) {
	id := c.Param("id")
	var document models.DocumentStaff

	if err := config.DB.Preload("User").First(&document, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"document_staff": document})
}

// ======================================================
// UPDATE DOCUMENT STAFF
// ======================================================
func UpdateDocumentStaff(c *gin.Context) {
	id := c.Param("id")
	var document models.DocumentStaff

	if err := config.DB.First(&document, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	// Update subject jika ada
	subject := c.PostForm("subject")
	if subject != "" {
		document.Subject = subject
	}

	// Cek file baru
	fileHeader, err := c.FormFile("file")
	if err == nil {
		src, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tidak dapat membuka file"})
			return
		}
		defer src.Close()

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		var resourceType string
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".webp":
			resourceType = "image"
		case ".pdf":
			resourceType = "raw"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung. Hanya PDF atau gambar"})
			return
		}

		url, publicID, err := config.UploadToCloudinary(src, resourceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload gagal: " + err.Error()})
			return
		}

		// Hapus file lama jika ada
		if document.PublicID != "" {
			config.DeleteFromCloudinary(document.PublicID, document.ResourceType)
		}

		document.FileName = url
		document.PublicID = publicID
		document.ResourceType = resourceType
	}

	if err := config.DB.Save(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan perubahan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil diperbarui", "document_staff": document})
}

// ======================================================
// DELETE DOCUMENT STAFF
// ======================================================
func DeleteDocumentStaff(c *gin.Context) {
	id := c.Param("id")
	var document models.DocumentStaff

	if err := config.DB.First(&document, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	// Hapus file dari Cloudinary
	config.DeleteFromCloudinary(document.PublicID, document.ResourceType)

	// Hapus dari DB
	config.DB.Delete(&document)

	c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
}
