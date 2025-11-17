package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func DocumentRoutes(r *gin.RouterGroup) {
	api := r.Group("/documents")

	// WAJIB login + wajib admin
	api.Use(middleware.AuthMiddleware(), middleware.AdminOnly())

	{
		api.POST("/", controllers.CreateDocument)
		api.GET("/", controllers.GetDocuments)
		api.GET("/:id", controllers.GetDocumentByID)
		api.PUT("/:id", controllers.UpdateDocument)
		api.DELETE("/:id", controllers.DeleteDocument)
	}
}
