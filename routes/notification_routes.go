package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func NotificationRoutes(router *gin.RouterGroup) {
	// Group sudah punya prefix "/api" dari main.go
	// Jadi route ini akan jadi "/api/notifications"

	notifications := router.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware())
	{
		// GET /api/notifications
		notifications.GET("", controllers.GetNotifications)  // Tanpa slash
		notifications.GET("/", controllers.GetNotifications) // Dengan slash (fallback)

		// POST /api/notifications/:id/read
		notifications.POST("/:id/read", controllers.MarkNotificationAsRead)
	}
}
