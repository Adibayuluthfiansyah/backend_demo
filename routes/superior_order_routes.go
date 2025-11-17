package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func SuperiorOrderRoutes(router *gin.RouterGroup) {
	superior := router.Group("/superior_orders")

	// Semua route butuh login
	superior.Use(middleware.AuthMiddleware())

	{
		// Hanya admin yang boleh akses semua route CRUD
		// Handle tanpa trailing slash
		superior.POST("", middleware.AdminOnly(), controllers.CreateSuperiorOrder)
		superior.GET("", middleware.AdminOnly(), controllers.GetSuperiorOrders)

		// Handle dengan trailing slash (fallback)
		superior.POST("/", middleware.AdminOnly(), controllers.CreateSuperiorOrder)
		superior.GET("/", middleware.AdminOnly(), controllers.GetSuperiorOrders)

		// Routes dengan parameter
		superior.GET("/:id", middleware.AdminOnly(), controllers.GetSuperiorOrdersByDocument)
		superior.PUT("/:id", middleware.AdminOnly(), controllers.UpdateSuperiorOrder)
		superior.DELETE("/:id", middleware.AdminOnly(), controllers.DeleteSuperiorOrder)
	}
}
