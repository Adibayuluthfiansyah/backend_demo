package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func SuperiorOrderRoutes(r *gin.Engine) {
	superior := r.Group("/api/superior_orders")

	// Semua route butuh login
	superior.Use(middleware.AuthMiddleware())

	{
		// Hanya admin yang boleh akses semua route CRUD
		superior.POST("/", middleware.AdminOnly(), controllers.CreateSuperiorOrder)
		superior.GET("/", middleware.AdminOnly(), controllers.GetSuperiorOrders)
		superior.GET("/:id", middleware.AdminOnly(), controllers.GetSuperiorOrdersByDocument)
		superior.PUT("/:id", middleware.AdminOnly(), controllers.UpdateSuperiorOrder)
		superior.DELETE("/:id", middleware.AdminOnly(), controllers.DeleteSuperiorOrder)
	}
}
