package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")

	// Rute create admin tidak perlu auth
	users.POST("/admin", controllers.CreateAdmin) // nanti hapus pas deploy

	// Grup untuk rute yang perlu auth
	usersAuth := users.Group("")
	usersAuth.Use(middleware.AuthMiddleware())
	{
		// Handle tanpa trailing slash
		usersAuth.GET("", controllers.GetUsers)
		usersAuth.GET("/me", controllers.GetMe)
		usersAuth.GET("/:id", controllers.GetUserByID)
		usersAuth.PUT("/:id", controllers.UpdateUser)

		// Handle dengan trailing slash (fallback)
		usersAuth.GET("/", controllers.GetUsers)

		// Admin only routes
		usersAuth.POST("/staff", middleware.AdminOnly(), controllers.CreateStaff)
		usersAuth.DELETE("/:id", middleware.AdminOnly(), controllers.DeleteUser)
	}
}
