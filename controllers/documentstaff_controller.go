package controllers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/models"

	"github.com/gin-gonic/gin"
)

// ======================================================
// CREATE STAFF DOCUMENT
// ======================================================
func CreateDocumentStaff(c *gin.Context) {
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	subject := c.PostForm("subject")
	// Sender dan LetterType opsional dari Frontend
	sender := c.PostForm("sender")
	letterType := c.PostForm("letter_type")

	// Validasi Subject Wajib
	if subject == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subject wajib diisi"})
		return
	}

	// Default Value jika kosong (karena dihilangkan di frontend staff)
	if sender == "" {
		sender = user.Name // Default pengirim nama sendiri
	}
	if letterType == "" {
		letterType = "keluar" // Default jenis surat
	}

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

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file buffer"})
		return
	}
	reader := bytes.NewReader(fileBytes)

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	var resourceType string
	var folder string

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		resourceType = "image"
		folder = "dinsos_kuburaya/gambar"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		resourceType = "raw"
		folder = "dinsos_kuburaya/arsip"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung"})
		return
	}

	uploadResult, err := config.UploadToCloudinary(reader, fileHeader.Filename, folder, resourceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload gagal: " + err.Error()})
		return
	}

	document := models.DocumentStaff{
		UserID:       user.ID,
		Sender:       sender,
		Subject:      subject,
		LetterType:   letterType,
		FileName:     uploadResult.SecureURL,
		PublicID:     uploadResult.PublicID,
		ResourceType: resourceType,
	}

	if err := config.DB.Create(&document).Error; err != nil {
		config.DeleteFromCloudinary(uploadResult.PublicID, resourceType)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error: " + err.Error()})
		return
	}

	// Log Activity
	msg := fmt.Sprintf("Mengupload dokumen baru dengan subjek: %s", document.Subject)
	LogActivity(user.ID, user.Name, "UPLOAD_DOKUMEN", msg)

	// Notifikasi ke Admin
	go func() {
		var admins []models.User
		if err := config.DB.Where("role = ?", "admin").Find(&admins).Error; err == nil {
			for _, admin := range admins {
				msg := fmt.Sprintf("Staff %s baru saja mengupload dokumen: %s", user.Name, document.Subject)
				link := fmt.Sprintf("/dashboard/documents/%s", document.ID)
				_ = CreateNotification(admin.ID, msg, link)
			}
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Dokumen berhasil diupload",
		"document": document,
	})
}

// ======================================================
// GET ALL STAFF DOCUMENTS (GABUNGAN ADMIN & STAFF)
// ======================================================
func GetDocumentStaffs(c *gin.Context) {
	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))
	search := c.Query("search")
	letterType := c.Query("letter_type")

	// Struct khusus untuk menampung hasil gabungan tabel
	type CombinedDoc struct {
		ID         string  `json:"id"`
		Sender     string  `json:"sender"`
		Subject    string  `json:"subject"`
		LetterType string  `json:"letter_type"`
		FileName   string  `json:"file_name"`
		UserID     *string `json:"user_id"`
		UserName   string  `json:"user_name"`
		CreatedAt  string  `json:"created_at"`
		UpdatedAt  string  `json:"updated_at"`
		Source     string  `json:"source"` // Untuk membedakan asal dokumen di frontend
	}

	var combinedDocs []CombinedDoc

	// Filter Pencarian
	searchWhere := ""
	searchArgs := []interface{}{}
	if search != "" {
		searchPattern := "%" + search + "%"
		// Mencari di sender, subject, atau nama file
		searchWhere = "AND (sender LIKE ? OR subject LIKE ? OR file_name LIKE ?)"
		searchArgs = []interface{}{searchPattern, searchPattern, searchPattern}
	}

	// Filter Tipe Surat
	letterTypeWhere := ""
	letterTypeArgs := []interface{}{}
	if letterType != "" && letterType != "all" {
		letterTypeWhere = "AND letter_type = ?"
		letterTypeArgs = []interface{}{letterType}
	}

	// LOGIKA FILTER USER:
	// 1. Tabel `documents` (Admin Upload): TIDAK ADA filter user_id, karena staff boleh lihat semua dokumen admin.
	// 2. Tabel `document_staffs` (Staff Upload):
	//    - Jika Admin: Lihat semua.
	//    - Jika Staff: Hanya lihat milik sendiri.

	staffFilter := ""
	if user.Role != "admin" {
		staffFilter = fmt.Sprintf("AND user_id = '%s'", user.ID)
	}

	// QUERY GABUNGAN (UNION)
	query := fmt.Sprintf(`
		(
			SELECT 
				id, sender, subject, letter_type, file_url as file_name, 
				user_id, 'Admin' as user_name, created_at, updated_at, 'document' as source
			FROM documents
			WHERE 1=1 %s %s
		)
		UNION ALL
		(
			SELECT 
				id, sender, subject, letter_type, file_name, 
				user_id, 'Staff' as user_name, created_at, updated_at, 'document_staff' as source
			FROM document_staffs
			WHERE 1=1 %s %s %s
		)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, searchWhere, letterTypeWhere, searchWhere, letterTypeWhere, staffFilter)

	// Susun Arguments
	args := []interface{}{}
	// Args untuk query pertama (Documents Admin)
	args = append(args, searchArgs...)
	args = append(args, letterTypeArgs...)
	// Args untuk query kedua (Document Staffs)
	args = append(args, searchArgs...)
	args = append(args, letterTypeArgs...)
	// Args untuk Limit/Offset
	args = append(args, perPage, (page-1)*perPage)

	if err := config.DB.Raw(query, args...).Scan(&combinedDocs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil dokumen: " + err.Error()})
		return
	}

	// Hitung Total (Untuk Pagination Frontend)
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT id FROM documents WHERE 1=1 %s %s
			UNION ALL
			SELECT id FROM document_staffs WHERE 1=1 %s %s %s
		) as total_docs
	`, searchWhere, letterTypeWhere, searchWhere, letterTypeWhere, staffFilter)

	countArgs := []interface{}{}
	countArgs = append(countArgs, searchArgs...)
	countArgs = append(countArgs, letterTypeArgs...)
	countArgs = append(countArgs, searchArgs...)
	countArgs = append(countArgs, letterTypeArgs...)

	var total int64
	config.DB.Raw(countQuery, countArgs...).Count(&total)

	// Hitung Last Page
	lastPage := int(total) / perPage
	if int(total)%perPage != 0 {
		lastPage++
	}
	if lastPage == 0 && total > 0 {
		lastPage = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"documents":    combinedDocs,
		"total":        total,
		"current_page": page,
		"last_page":    lastPage,
		"per_page":     perPage,
	})
}

// ======================================================
// GET STAFF DOCUMENT BY ID (Mencari di kedua tabel)
// ======================================================
func GetDocumentStaffByID(c *gin.Context) {
	id := c.Param("id")

	// Cek Tabel Staff Dulu
	var docStaff models.DocumentStaff
	errStaff := config.DB.Preload("User").First(&docStaff, "id = ?", id).Error

	if errStaff == nil {
		// Ketemu di tabel staff
		c.JSON(http.StatusOK, gin.H{"document": docStaff})
		return
	}

	// Jika tidak, Cek Tabel Admin
	var doc models.Document
	errDoc := config.DB.Preload("User").First(&doc, "id = ?", id).Error

	if errDoc == nil {
		// Ketemu di tabel admin -> Konversi format response agar frontend staff tidak error
		response := gin.H{
			"id":          doc.ID,
			"sender":      doc.Sender,
			"subject":     doc.Subject,
			"letter_type": doc.LetterType,
			"file_name":   doc.FileURL, // Mapping FileURL ke file_name
			"user_id":     doc.UserID,
			"created_at":  doc.CreatedAt,
			"updated_at":  doc.UpdatedAt,
			"user":        doc.User,
			"source":      "document", // Penanda
		}
		c.JSON(http.StatusOK, gin.H{"document": response})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
}

// ======================================================
// UPDATE STAFF DOCUMENT
// ======================================================
func UpdateDocumentStaff(c *gin.Context) {
	id := c.Param("id")

	userRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userRaw.(models.User)

	var document models.DocumentStaff

	if err := config.DB.First(&document, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
		return
	}

	if user.Role != "admin" && document.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk mengedit dokumen ini"})
		return
	}

	sender := c.PostForm("sender")
	subject := c.PostForm("subject")
	letterType := c.PostForm("letter_type")

	updates := map[string]interface{}{}

	if sender != "" {
		updates["sender"] = sender
	}
	if subject != "" {
		updates["subject"] = subject
	}
	if letterType != "" {
		updates["letter_type"] = letterType
	}

	fileHeader, err := c.FormFile("file")
	if err == nil {
		src, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tidak dapat membuka file"})
			return
		}
		defer src.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file buffer"})
			return
		}
		reader := bytes.NewReader(fileBytes)

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		var resourceType string
		var folder string

		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".webp":
			resourceType = "image"
			folder = "dinsos_kuburaya/gambar"
		case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
			resourceType = "raw"
			folder = "dinsos_kuburaya/arsip"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung"})
			return
		}

		uploadResult, err := config.UploadToCloudinary(reader, fileHeader.Filename, folder, resourceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload gagal: " + err.Error()})
			return
		}

		if document.PublicID != "" {
			config.DeleteFromCloudinary(document.PublicID, document.ResourceType)
		}
		updates["file_name"] = uploadResult.SecureURL
		updates["public_id"] = uploadResult.PublicID
		updates["resource_type"] = resourceType
	}

	if len(updates) > 0 {
		if err := config.DB.Model(&document).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan perubahan"})
			return
		}
	}

	config.DB.Preload("User").Find(&document)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Dokumen berhasil diperbarui",
		"document": document,
	})
}

// ======================================================
// DELETE DOCUMENT STAFF
// ======================================================
func DeleteDocumentStaff(c *gin.Context) {
	id := c.Param("id")

	var docStaff models.DocumentStaff
	errStaff := config.DB.First(&docStaff, "id = ?", id).Error

	if errStaff == nil {
		if docStaff.PublicID != "" {
			config.DeleteFromCloudinary(docStaff.PublicID, docStaff.ResourceType)
		}
		config.DB.Delete(&docStaff)
		c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
		return
	}

	// Jika admin menghapus dokumen via endpoint ini
	var doc models.Document
	errDoc := config.DB.First(&doc, "id = ?", id).Error

	if errDoc == nil {
		if doc.PublicID != "" {
			config.DeleteFromCloudinary(doc.PublicID, doc.ResourceType)
		}
		config.DB.Delete(&doc)
		c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
}

// ======================================================
// DOWNLOAD DOCUMENT STAFF
// ======================================================
func DownloadDocumentStaff(c *gin.Context) {
	id := c.Param("id")

	var docStaff models.DocumentStaff
	errStaff := config.DB.First(&docStaff, "id = ?", id).Error

	if errStaff == nil && docStaff.FileName != "" {
		c.Redirect(http.StatusTemporaryRedirect, docStaff.FileName)
		return
	}

	var doc models.Document
	errDoc := config.DB.First(&doc, "id = ?", id).Error

	if errDoc == nil && doc.FileURL != "" {
		c.Redirect(http.StatusTemporaryRedirect, doc.FileURL)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File tidak tersedia"})
}
