package controllers

import (
	"net/http"
	"path/filepath"
	"strings"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

// =======================
// CREATE DOCUMENT (Cloudinary Integration)
// =======================
func CreateDocument(c *gin.Context) {
	sender := c.PostForm("sender")
	subject := c.PostForm("subject")
	letterType := c.PostForm("letter_type")
	userID := c.PostForm("user_id")

	// Ambil file dari form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan"})
		return
	}

	// Buka file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file: " + err.Error()})
		return
	}
	defer src.Close()

	// ======================
	// Deteksi tipe file
	// ======================
	ext := strings.ToLower(filepath.Ext(file.Filename))
	resourceType := "raw"
	folder := "arsip"

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		resourceType = "image"
		folder = "gambar"
	}

	// ======================
	// Upload ke Cloudinary
	// ======================
	fileURL, err := config.UploadToCloudinary(src, file.Filename, folder, resourceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke Cloudinary: " + err.Error()})
		return
	}

	// ======================
	// Simpan ke database
	// ======================
	document := models.Document{
		Sender:     sender,
		FileName:   fileURL, // simpan URL hasil upload, bukan nama file lokal
		Subject:    subject,
		LetterType: letterType,
	}

	if userID != "" {
		document.UserID = &userID
	}

	if err := config.DB.Create(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan dokumen: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Dokumen berhasil diupload dan disimpan",
		"file_url":  fileURL,
		"folder":    folder,
		"file_type": resourceType,
		"document":  document,
	})
}

// =======================
// GET ALL DOCUMENTS
// =======================
func GetDocuments(c *gin.Context) {
	var documents []models.Document

	if err := config.DB.Preload("User").Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data dokumen: " + err.Error()})
		return
	}

	// Bentuk respons agar menampilkan nama user dengan jelas
	var response []gin.H
	for _, doc := range documents {
		response = append(response, gin.H{
			"id":          doc.ID,
			"sender":      doc.Sender,
			"file_name":   doc.FileName,
			"subject":     doc.Subject,
			"letter_type": doc.LetterType,
			"user_id":     doc.UserID,
			"user_name":   doc.User.Name, // ambil nama user dari relasi
			"created_at":  doc.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": response,
	})
}

// =======================
// GET DOCUMENT BY ID
// =======================
func GetDocumentByID(c *gin.Context) {
	id := c.Param("id")
	var document models.Document

	if err := config.DB.Preload("User").First(&document, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	response := gin.H{
		"id":          document.ID,
		"sender":      document.Sender,
		"file_name":   document.FileName,
		"subject":     document.Subject,
		"letter_type": document.LetterType,
		"user_id":     document.UserID,
		"user_name":   document.User.Name, // tampilkan nama uploader
		"created_at":  document.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"document": response})
}

// =======================
// UPDATE DOCUMENT
// =======================
func UpdateDocument(c *gin.Context) {
	id := c.Param("id")
	var document models.Document

	if err := config.DB.First(&document, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	var updatedData models.Document
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	document.Sender = updatedData.Sender
	document.FileName = updatedData.FileName
	document.Subject = updatedData.Subject
	document.LetterType = updatedData.LetterType
	document.UserID = updatedData.UserID

	if err := config.DB.Save(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui dokumen: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Dokumen berhasil diperbarui",
		"document": document,
	})
}

// =======================
// DELETE DOCUMENT
// =======================
func DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	var document models.Document

	// gunakan "id = ?" agar cocok dengan UUID (string)
	if err := config.DB.First(&document, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus dokumen: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
}
