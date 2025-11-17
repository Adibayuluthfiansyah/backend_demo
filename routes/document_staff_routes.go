package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func DocumentStaffRoutes(r *gin.RouterGroup) {
	docStaff := r.Group("/document_staff")
	docStaff.Use(middleware.AuthMiddleware())
	{
		// Handle tanpa trailing slash
		docStaff.POST("", controllers.CreateDocumentStaff)
		docStaff.GET("", controllers.GetDocumentStaffs)

		// Handle dengan trailing slash (fallback)
		docStaff.POST("/", controllers.CreateDocumentStaff)
		docStaff.GET("/", controllers.GetDocumentStaffs)

		docStaff.GET("/:id", controllers.GetDocumentStaffByID)
		docStaff.PUT("/:id", controllers.UpdateDocumentStaff)
		docStaff.DELETE("/:id", controllers.DeleteDocumentStaff)
	}
}
