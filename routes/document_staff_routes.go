package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func DocumentStaffRoutes(r *gin.RouterGroup) {
	docStaff := r.Group("/document_staff")

	// Semua route harus login
	docStaff.Use(middleware.AuthMiddleware())

	{
		// Semua user yang login (staff & admin) bisa akses
		docStaff.POST("/", controllers.CreateDocumentStaff)
		docStaff.GET("/", controllers.GetDocumentStaffs)
		docStaff.GET("/:id", controllers.GetDocumentStaffByID)
		docStaff.PUT("/:id", controllers.UpdateDocumentStaff)
		docStaff.DELETE("/:id", controllers.DeleteDocumentStaff)
	}
}
