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
// GET DOCUMENT BY ID (CEK DI KEDUA TABEL + PERMISSION)
// =======================
func GetDocumentByID(c *gin.Context) {
	id := c.Param("id")

	// Ambil user dari context
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	// 1. Coba cari di documents dulu
	var document models.Document
	errDoc := config.DB.Preload("User").Where("id = ?", id).First(&document).Error

	if errDoc == nil {
		// Ketemu di documents

		// ✅ CEK PERMISSION: Staff hanya bisa lihat dokumen miliknya
		if user.Role != "admin" {
			if document.UserID == nil || *document.UserID != user.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses ke dokumen ini"})
				return
			}
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
		return
	}

	// 2. Kalau tidak ketemu, coba cari di document_staffs
	var docStaff models.DocumentStaff
	errStaff := config.DB.Preload("User").First(&docStaff, "id = ?", id).Error

	if errStaff == nil {
		// Ketemu di document_staffs

		// ✅ CEK PERMISSION: Staff hanya bisa lihat dokumen miliknya
		if user.Role != "admin" && docStaff.UserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses ke dokumen ini"})
			return
		}

		userName := "-"
		if docStaff.User.Name != "" {
			userName = docStaff.User.Name
		}

		response := gin.H{
			"id":          docStaff.ID,
			"sender":      docStaff.Sender,
			"file_name":   docStaff.FileName,
			"file_url":    docStaff.FileName,
			"subject":     docStaff.Subject,
			"letter_type": docStaff.LetterType,
			"user_id":     docStaff.UserID,
			"user_name":   userName,
			"created_at":  docStaff.CreatedAt,
			"updated_at":  docStaff.UpdatedAt,
		}

		c.JSON(http.StatusOK, gin.H{"document": response})
		return
	}

	// 3. Tidak ketemu di keduanya
	c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
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
// DELETE DOCUMENT (CEK DI KEDUA TABEL)
// =======================
func DeleteDocument(c *gin.Context) {
	id := c.Param("id")

	// 1. Coba hapus dari documents
	var document models.Document
	errDoc := config.DB.Where("id = ?", id).First(&document).Error

	if errDoc == nil {
		// Hapus dari Cloudinary
		if document.PublicID != "" && document.ResourceType != "" {
			err := config.DeleteFromCloudinary(document.PublicID, document.ResourceType)
			if err != nil {
				fmt.Printf("⚠️ Gagal menghapus file dari Cloudinary: %v\n", err)
			}
		}

		// Hapus dari database
		if err := config.DB.Delete(&document).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus dokumen: " + err.Error()})
			return
		}

		// Log activity
		if userRaw, exists := c.Get("user"); exists {
			actor := userRaw.(models.User)
			CreateActivityLog(actor.ID, actor.Name, "DELETE_DOCUMENT", "Menghapus dokumen: "+document.FileName)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
		return
	}

	// 2. Coba hapus dari document_staffs
	var docStaff models.DocumentStaff
	errStaff := config.DB.First(&docStaff, "id = ?", id).Error

	if errStaff == nil {
		// Hapus dari Cloudinary
		if docStaff.PublicID != "" {
			config.DeleteFromCloudinary(docStaff.PublicID, docStaff.ResourceType)
		}

		// Hapus dari database
		config.DB.Delete(&docStaff)

		c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
		return
	}

	// 3. Tidak ketemu
	c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
}

func DownloadDocument(c *gin.Context) {
	id := c.Param("id")

	// Ambil user
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	// Cari di documents
	var document models.Document
	errDoc := config.DB.First(&document, "id = ?", id).Error

	if errDoc == nil && document.FileURL != "" {
		// ✅ CEK PERMISSION
		if user.Role != "admin" {
			if document.UserID == nil || *document.UserID != user.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Tidak memiliki akses"})
				return
			}
		}

		c.Redirect(http.StatusTemporaryRedirect, document.FileURL)
		return
	}

	// Cari di document_staffs
	var docStaff models.DocumentStaff
	errStaff := config.DB.First(&docStaff, "id = ?", id).Error

	if errStaff == nil && docStaff.FileName != "" {
		// ✅ CEK PERMISSION
		if user.Role != "admin" && docStaff.UserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tidak memiliki akses"})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, docStaff.FileName)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File tidak tersedia"})
}
