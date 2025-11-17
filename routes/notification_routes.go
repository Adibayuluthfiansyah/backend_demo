package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func NotificationRoutes(router *gin.RouterGroup) {
	notifications := router.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware())
	{
		// Handle tanpa trailing slash
		notifications.GET("", controllers.GetNotifications)

		// Handle dengan trailing slash (fallback)
		notifications.GET("/", controllers.GetNotifications)

		notifications.POST("/:id/read", controllers.MarkNotificationAsRead)
	}
}
