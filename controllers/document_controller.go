package controllers

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

// =======================
// CREATE DOCUMENT
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

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file"})
		return
	}
	defer src.Close()

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
		config.DeleteFromCloudinary(uploadResult.PublicID, resourceType)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan dokumen: " + err.Error()})
		return
	}

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
			"user":        doc.User,
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
// GET DOCUMENT BY ID (FIXED PERMISSION FOR STAFF)
// =======================
func GetDocumentByID(c *gin.Context) {
	id := c.Param("id")
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	var document models.Document
	errDoc := config.DB.Preload("User").Where("id = ?", id).First(&document).Error

	if errDoc == nil {
		if user.Role != "admin" {
			if document.UserID == nil || *document.UserID != user.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses ke dokumen admin ini"})
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
			"user":        document.User,
		}
		c.JSON(http.StatusOK, gin.H{"document": response})
		return
	}

	var docStaff models.DocumentStaff
	errStaff := config.DB.Preload("User").First(&docStaff, "id = ?", id).Error

	if errStaff == nil {
		if user.Role != "admin" && docStaff.UserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses ke dokumen staff lain"})
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
			"user":        docStaff.User,
			"source":      "staff",
		}
		c.JSON(http.StatusOK, gin.H{"document": response})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
}

// =======================
// UPDATE DOCUMENT (SUPPORT STAFF UPDATE)
// =======================
func UpdateDocument(c *gin.Context) {
	id := c.Param("id")

	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	var document models.Document
	if err := config.DB.Where("id = ?", id).First(&document).Error; err == nil {
		if user.Role != "admin" {
			if document.UserID == nil || *document.UserID != user.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Tidak punya akses edit"})
				return
			}
		}
		updateDocLogic(c, &document, user, "documents")
		return
	}

	var docStaff models.DocumentStaff
	if err := config.DB.Where("id = ?", id).First(&docStaff).Error; err == nil {
		if user.Role != "admin" && docStaff.UserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tidak punya akses edit"})
			return
		}
		updateDocStaffLogic(c, &docStaff, user)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
}

// Helper untuk update tabel Document Admin
func updateDocLogic(c *gin.Context, document *models.Document, user models.User, table string) {
	sender := c.PostForm("sender")
	subject := c.PostForm("subject")
	letterType := c.PostForm("letter_type")

	if sender != "" {
		document.Sender = sender
	}
	if subject != "" {
		document.Subject = subject
	}
	if letterType != "" {
		document.LetterType = letterType
	}

	// Handle File Upload
	fileHeader, err := c.FormFile("file")
	if err == nil {
		src, _ := fileHeader.Open()
		defer src.Close()
		fileBytes, _ := io.ReadAll(src)
		reader := bytes.NewReader(fileBytes)

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		resourceType := "raw"
		folder := "arsip"
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
			resourceType = "image"
			folder = "gambar"
		}

		if document.PublicID != "" {
			config.DeleteFromCloudinary(document.PublicID, document.ResourceType)
		}

		uploadResult, err := config.UploadToCloudinary(reader, fileHeader.Filename, folder, resourceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload file: " + err.Error()})
			return
		}

		document.FileName = fileHeader.Filename
		document.FileURL = uploadResult.SecureURL
		document.PublicID = uploadResult.PublicID
		document.ResourceType = uploadResult.ResourceType
	}

	if err := config.DB.Save(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan update: " + err.Error()})
		return
	}

	CreateActivityLog(user.ID, user.Name, "UPDATE_DOCUMENT", "Memperbarui dokumen: "+document.Subject)
	c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil diperbarui", "document": document})
}

// Helper untuk update tabel DocumentStaff Staff
func updateDocStaffLogic(c *gin.Context, docStaff *models.DocumentStaff, user models.User) {
	sender := c.PostForm("sender")
	subject := c.PostForm("subject")
	letterType := c.PostForm("letter_type")

	if sender != "" {
		docStaff.Sender = sender
	}
	if subject != "" {
		docStaff.Subject = subject
	}
	if letterType != "" {
		docStaff.LetterType = letterType
	}

	fileHeader, err := c.FormFile("file")
	if err == nil {
		src, _ := fileHeader.Open()
		defer src.Close()
		fileBytes, _ := io.ReadAll(src)
		reader := bytes.NewReader(fileBytes)

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		resourceType := "raw"
		folder := "arsip"
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
			resourceType = "image"
			folder = "gambar"
		}

		if docStaff.PublicID != "" {
			config.DeleteFromCloudinary(docStaff.PublicID, docStaff.ResourceType)
		}

		uploadResult, err := config.UploadToCloudinary(reader, fileHeader.Filename, folder, resourceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload file: " + err.Error()})
			return
		}

		docStaff.FileName = uploadResult.SecureURL
		docStaff.PublicID = uploadResult.PublicID
		docStaff.ResourceType = uploadResult.ResourceType
	}

	if err := config.DB.Save(&docStaff).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan update staff: " + err.Error()})
		return
	}

	CreateActivityLog(user.ID, user.Name, "UPDATE_DOCUMENT", "Memperbarui dokumen staff: "+docStaff.Subject)
	c.JSON(http.StatusOK, gin.H{"message": "Dokumen staff berhasil diperbarui", "document": docStaff})
}

// =======================
// DELETE DOCUMENT
// =======================
func DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	var document models.Document
	if err := config.DB.Where("id = ?", id).First(&document).Error; err == nil {
		if user.Role != "admin" {
			if document.UserID == nil || *document.UserID != user.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Tidak punya akses hapus"})
				return
			}
		}

		if document.PublicID != "" {
			config.DeleteFromCloudinary(document.PublicID, document.ResourceType)
		}
		config.DB.Delete(&document)
		CreateActivityLog(user.ID, user.Name, "DELETE_DOCUMENT", "Menghapus dokumen: "+document.FileName)
		c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
		return
	}

	var docStaff models.DocumentStaff
	if err := config.DB.First(&docStaff, "id = ?", id).Error; err == nil {
		if user.Role != "admin" && docStaff.UserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tidak punya akses hapus"})
			return
		}

		if docStaff.PublicID != "" {
			config.DeleteFromCloudinary(docStaff.PublicID, docStaff.ResourceType)
		}
		config.DB.Delete(&docStaff)
		CreateActivityLog(user.ID, user.Name, "DELETE_DOCUMENT", "Menghapus dokumen staff: "+docStaff.FileName)
		c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
}

func DownloadDocument(c *gin.Context) {
	id := c.Param("id")
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	var document models.Document
	if err := config.DB.First(&document, "id = ?", id).Error; err == nil {
		if user.Role != "admin" {
			if document.UserID == nil || *document.UserID != user.ID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Tidak memiliki akses"})
				return
			}
		}
		c.Redirect(http.StatusTemporaryRedirect, document.FileURL)
		return
	}

	var docStaff models.DocumentStaff
	if err := config.DB.First(&docStaff, "id = ?", id).Error; err == nil {
		if user.Role != "admin" && docStaff.UserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tidak memiliki akses"})
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, docStaff.FileName)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File tidak tersedia"})
}
