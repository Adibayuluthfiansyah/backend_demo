package controllers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

// =======================
// CREATE DOCUMENT (FIXED)
// =======================
func CreateDocument(c *gin.Context) {
	sender := c.PostForm("sender")
	subject := c.PostForm("subject")
	letterType := c.PostForm("letter_type")

	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast user"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan"})
		return
	}

	// OPEN FILE
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file"})
		return
	}
	defer src.Close()

	// =======================================================
	// BACA FILE KE BUFFER AGAR TIDAK CORRUPT
	// =======================================================
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file buffer"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	resourceType := "raw"
	folder := "arsip"

	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
		resourceType = "image"
		folder = "gambar"
	} else if ext == ".pdf" {
		resourceType = "raw"
		folder = "arsip"
	}

	fmt.Printf("Upload Cloudinary | File: %s | Type: %s | Folder: %s\n",
		fileHeader.Filename, resourceType, folder)

	// Kirim buffer ke Cloudinary
	reader := bytes.NewReader(fileBytes)

	uploadResult, err := config.UploadToCloudinary(reader, fileHeader.Filename, folder, resourceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cloudinary upload gagal: " + err.Error()})
		return
	}

	userID := user.ID

	document := models.Document{
		Sender:       sender,
		FileName:     fileHeader.Filename,
		FileURL:      uploadResult.SecureURL,
		Subject:      subject,
		LetterType:   letterType,
		UserID:       &userID,
		PublicID:     uploadResult.PublicID,
		ResourceType: uploadResult.ResourceType,
	}

	if err := config.DB.Create(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan dokumen: " + err.Error()})
		return
	}

	// ===> TAMBAHAN: CATAT UPLOAD <===
	CreateActivityLog(user.ID, user.Name, "UPLOAD_DOCUMENT", "Mengunggah dokumen: "+document.FileName)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Dokumen berhasil diupload",
		"document": document,
		"file_url": uploadResult.SecureURL,
	})
}

// =======================
// GET ALL DOCUMENTS
// =======================
func GetDocuments(c *gin.Context) {
	var documents []models.Document

	search := c.Query("search")
	letterType := c.Query("letter_type")

	query := config.DB.Preload("User")

	if letterType != "" && letterType != "all" {
		query = query.Where("letter_type = ?", letterType)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("sender LIKE ? OR subject LIKE ? OR file_name LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if err := query.Order("created_at DESC").Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data dokumen: " + err.Error()})
		return
	}

	var response []gin.H
	for _, doc := range documents {
		userName := "-"
		if doc.User.Name != "" {
			userName = doc.User.Name
		}

		response = append(response, gin.H{
			"id":          doc.ID,
			"sender":      doc.Sender,
			"file_name":   doc.FileName,
			"file_url":    doc.FileURL,
			"subject":     doc.Subject,
			"letter_type": doc.LetterType,
			"user_id":     doc.UserID,
			"user_name":   userName,
			"created_at":  doc.CreatedAt,
			"updated_at":  doc.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"documents":    response,
		"total":        len(response),
		"current_page": 1,
		"last_page":    1,
		"per_page":     len(response),
	})
}

// =======================
// GET DOCUMENT BY ID
// =======================
func GetDocumentByID(c *gin.Context) {
	id := c.Param("id")
	var document models.Document

	if err := config.DB.Preload("User").Where("id = ?", id).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	userName := "-"
	if document.User.Name != "" {
		userName = document.User.Name
	}

	response := gin.H{
		"id":          document.ID,
		"sender":      document.Sender,
		"file_name":   document.FileName,
		"file_url":    document.FileURL,
		"subject":     document.Subject,
		"letter_type": document.LetterType,
		"user_id":     document.UserID,
		"user_name":   userName,
		"created_at":  document.CreatedAt,
		"updated_at":  document.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"document": response})
}

// =======================
// UPDATE DOCUMENT
// =======================
func UpdateDocument(c *gin.Context) {
	id := c.Param("id")
	var document models.Document

	if err := config.DB.Where("id = ?", id).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	var updatedData struct {
		Sender     string `json:"sender"`
		Subject    string `json:"subject"`
		LetterType string `json:"letter_type"`
	}

	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	document.Sender = updatedData.Sender
	document.Subject = updatedData.Subject
	document.LetterType = updatedData.LetterType

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

	if err := config.DB.Where("id = ?", id).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	// --- INI  UNTUK HAPUS DARI CLOUDINARY ---

	if document.PublicID != "" && document.ResourceType != "" {
		// Panggil fungsi DeleteFromCloudinary
		err := config.DeleteFromCloudinary(document.PublicID, document.ResourceType)
		if err != nil {
			fmt.Printf(" Gagal menghapus file dari Cloudinary (lanjutkan proses): %v\n", err)
		}
	} else {
		fmt.Printf("PublicID atau ResourceType kosong untuk dokumen %s, skip delete Cloudinary\n", id)
	}

	// Hapus dari database SETELAH mencoba hapus dari Cloudinary
	if err := config.DB.Delete(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus dokumen dari database: " + err.Error()})
		return
	}
	if userRaw, exists := c.Get("user"); exists {
		actor := userRaw.(models.User)
		CreateActivityLog(actor.ID, actor.Name, "DELETE_DOCUMENT", "Menghapus dokumen: "+document.FileName)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
}

// =======================
// DOWNLOAD DOCUMENT
// =======================
func DownloadDocument(c *gin.Context) {
	id := c.Param("id")
	var document models.Document

	if err := config.DB.First(&document, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	if document.FileURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link file tidak tersedia"})
		return
	}

	fmt.Printf("📥 Download request redirect ke: %s\n", document.FileURL)

	c.Redirect(http.StatusTemporaryRedirect, document.FileURL)
}
