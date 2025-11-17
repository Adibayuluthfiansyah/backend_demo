package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.RouterGroup) { // <-- Terima RouterGroup
	users := router.Group("/users") // <-- HANYA tambahkan /users

	// Rute create admin tidak perlu auth, jadi kita pindah
	users.POST("/admin", controllers.CreateAdmin)

	// Grup baru untuk rute yang perlu auth
	usersAuth := users.Group("")
	usersAuth.Use(middleware.AuthMiddleware())
	{
		usersAuth.POST("/staff", middleware.AdminOnly(), controllers.CreateStaff)
		usersAuth.DELETE("/:id", middleware.AdminOnly(), controllers.DeleteUser)
		usersAuth.GET("/", controllers.GetUsers)
		usersAuth.GET("/:id", controllers.GetUserByID)
		usersAuth.PUT("/:id", controllers.UpdateUser)

		// Rute /me dari perbaikan sebelumnya
		usersAuth.GET("/me", controllers.GetMe)

	}
}
