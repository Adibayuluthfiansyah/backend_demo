package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	users := r.Group("/api/users")

	users.POST("/admin", controllers.CreateAdmin)

	// Semua route butuh login
	users.Use(middleware.AuthMiddleware())

	{
		// Hanya admin yang boleh create staff dan delete user
		users.POST("/staff", middleware.AdminOnly(), controllers.CreateStaff)
		users.DELETE("/:id", middleware.AdminOnly(), controllers.DeleteUser)

		// Semua user (admin & staff) bisa melihat
		users.GET("/", controllers.GetUsers)
		users.GET("/:id", controllers.GetUserByID)

		// Update user bisa kamu tentukan (saya biarkan bebas)
		users.PUT("/:id", controllers.UpdateUser)
	}
}
