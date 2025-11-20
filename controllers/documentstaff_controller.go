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

	sender := c.PostForm("sender")
	subject := c.PostForm("subject")
	letterType := c.PostForm("letter_type")

	if sender == "" || subject == "" || letterType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sender, subject, dan letter_type wajib diisi"})
		return
	}

	if letterType != "masuk" && letterType != "keluar" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "letter_type harus 'masuk' atau 'keluar'"})
		return
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

	// Upload ke Cloudinary
	uploadResult, err := config.UploadToCloudinary(reader, fileHeader.Filename, folder, resourceType)
	if err != nil {
		fmt.Println("Upload Error:", err) // Debug print
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload gagal: " + err.Error()})
		return
	}

	// Simpan ke Database
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

	// CATAT LOG AKTIVITAS
	msg := fmt.Sprintf("Mengupload dokumen baru dengan subjek: %s", document.Subject)
	LogActivity(user.ID, user.Name, "UPLOAD_DOKUMEN", msg)

	// ============================================================
	// NOTIFIKASI KE ADMIN
	// ============================================================
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
	config.DB.Preload("User").Find(&document)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Dokumen berhasil diupload",
		"document": document,
	})
}

// ======================================================
// GET ALL STAFF DOCUMENTS (ADMIN + STAFF)
// ======================================================
func GetDocumentStaffs(c *gin.Context) {
	// 1. Ambil User
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

	// ======================================================
	// ADMIN: Gabungkan documents + document_staffs dengan UNION
	// ======================================================
	if user.Role == "admin" {
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
			Source     string  `json:"source"`
		}

		var combinedDocs []CombinedDoc

		searchWhere := ""
		searchArgs := []interface{}{}
		if search != "" {
			searchPattern := "%" + search + "%"
			searchWhere = "AND (d.sender LIKE ? OR d.subject LIKE ? OR d.file_name LIKE ? OR u.name LIKE ?)"
			searchArgs = []interface{}{searchPattern, searchPattern, searchPattern, searchPattern}
		}

		letterTypeWhere := ""
		letterTypeArgs := []interface{}{}
		if letterType != "" && letterType != "all" {
			letterTypeWhere = "AND d.letter_type = ?"
			letterTypeArgs = []interface{}{letterType}
		}

		query := fmt.Sprintf(`
			SELECT 
				d.id, 
				d.sender, 
				d.subject, 
				d.letter_type, 
				d.file_url as file_name, 
				d.user_id,
				COALESCE(u.name, '-') as user_name,
				d.created_at, 
				d.updated_at, 
				'document' as source
			FROM documents d
			LEFT JOIN users u ON d.user_id = u.id
			WHERE 1=1 %s %s
			
			UNION ALL
			
			SELECT 
				ds.id, 
				ds.sender, 
				ds.subject, 
				ds.letter_type, 
				ds.file_name, 
				ds.user_id,
				COALESCE(u2.name, '-') as user_name,
				ds.created_at, 
				ds.updated_at, 
				'document_staff' as source
			FROM document_staffs ds
			LEFT JOIN users u2 ON ds.user_id = u2.id
			WHERE 1=1 %s %s
			
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`, searchWhere, letterTypeWhere, searchWhere, letterTypeWhere)

		args := []interface{}{}
		args = append(args, searchArgs...)
		args = append(args, letterTypeArgs...)
		args = append(args, searchArgs...)
		args = append(args, letterTypeArgs...)
		args = append(args, perPage, (page-1)*perPage)

		// Execute query
		err := config.DB.Raw(query, args...).Scan(&combinedDocs).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil dokumen: " + err.Error()})
			return
		}

		// Hitung total
		countQuery := fmt.Sprintf(`
			SELECT COUNT(*) as total FROM (
				SELECT d.id 
				FROM documents d
				LEFT JOIN users u ON d.user_id = u.id
				WHERE 1=1 %s %s
				
				UNION ALL
				
				SELECT ds.id 
				FROM document_staffs ds
				LEFT JOIN users u2 ON ds.user_id = u2.id
				WHERE 1=1 %s %s
			) as combined
		`, searchWhere, letterTypeWhere, searchWhere, letterTypeWhere)

		countArgs := []interface{}{}
		countArgs = append(countArgs, searchArgs...)
		countArgs = append(countArgs, letterTypeArgs...)
		countArgs = append(countArgs, searchArgs...)
		countArgs = append(countArgs, letterTypeArgs...)

		var total int64
		config.DB.Raw(countQuery, countArgs...).Count(&total)

		// Hitung last page
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
		return
	}

	// ======================================================
	// STAFF hanya ambil document staff
	// ======================================================
	var documents []models.DocumentStaff

	query := config.DB.Model(&models.DocumentStaff{}).Where("user_id = ?", user.ID)

	if search != "" {
		searchQuery := "%" + search + "%"
		query = query.Where(
			"sender LIKE ? OR subject LIKE ? OR file_name LIKE ?",
			searchQuery, searchQuery, searchQuery,
		)
	}

	if letterType != "" && letterType != "all" {
		query = query.Where("letter_type = ?", letterType)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * perPage
	err := query.
		Offset(offset).
		Limit(perPage).
		Order("created_at DESC").
		Preload("User").
		Find(&documents).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil dokumen"})
		return
	}

	lastPage := int(total) / perPage
	if int(total)%perPage != 0 {
		lastPage++
	}
	if lastPage == 0 && total > 0 {
		lastPage = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"documents":    documents,
		"total":        total,
		"current_page": page,
		"last_page":    lastPage,
		"per_page":     perPage,
	})
}

// ======================================================
// GET BY ID Cek di KEDUA tabel
// ======================================================
func GetDocumentStaffByID(c *gin.Context) {
	id := c.Param("id")

	var docStaff models.DocumentStaff
	errStaff := config.DB.Preload("User").First(&docStaff, "id = ?", id).Error

	if errStaff == nil {
		c.JSON(http.StatusOK, gin.H{"document": docStaff})
		return
	}

	var doc models.Document
	errDoc := config.DB.Preload("User").First(&doc, "id = ?", id).Error

	if errDoc == nil {
		response := gin.H{
			"id":          doc.ID,
			"sender":      doc.Sender,
			"subject":     doc.Subject,
			"letter_type": doc.LetterType,
			"file_name":   doc.FileURL,
			"user_id":     doc.UserID,
			"created_at":  doc.CreatedAt,
			"updated_at":  doc.UpdatedAt,
			"user":        doc.User,
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
		if letterType != "masuk" && letterType != "keluar" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "letter_type harus 'masuk' atau 'keluar'"})
			return
		}
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

		// Upload to Cloudinary
		uploadResult, err := config.UploadToCloudinary(reader, fileHeader.Filename, folder, resourceType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload gagal: " + err.Error()})
			return
		}

		// Delete data lama jika ada
		if document.PublicID != "" {
			err := config.DeleteFromCloudinary(document.PublicID, document.ResourceType)
			if err != nil {
				fmt.Println("Warning: Gagal menghapus file lama:", err)
			}
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
// DELETE Cek di KEDUA tabel
// ======================================================
func DeleteDocumentStaff(c *gin.Context) {
	id := c.Param("id")

	var docStaff models.DocumentStaff
	errStaff := config.DB.First(&docStaff, "id = ?", id).Error

	if errStaff == nil {
		// Hapus file dari Cloudinary
		if docStaff.PublicID != "" {
			config.DeleteFromCloudinary(docStaff.PublicID, docStaff.ResourceType)
		}
		config.DB.Delete(&docStaff)
		c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
		return
	}

	var doc models.Document
	errDoc := config.DB.First(&doc, "id = ?", id).Error

	if errDoc == nil {
		if doc.PublicID != "" {
			config.DeleteFromCloudinary(doc.PublicID, doc.ResourceType)
		}
		config.DB.Delete(&doc)

		if userRaw, exists := c.Get("user"); exists {
			actor := userRaw.(models.User)
			CreateActivityLog(actor.ID, actor.Name, "DELETE_DOCUMENT", "Menghapus dokumen: "+doc.FileName)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Dokumen berhasil dihapus"})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Dokumen tidak ditemukan"})
}

// ======================================================
// DOWNLOAD Cek di KEDUA tabel
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
