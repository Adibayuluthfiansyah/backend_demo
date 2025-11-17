package routes

import (
	"dinsos_kuburaya/controllers"
	"dinsos_kuburaya/middleware"

	"github.com/gin-gonic/gin"
)

func LogoutRoutes(r *gin.RouterGroup) {
	{
		r.POST("/logout", middleware.AuthMiddleware(), controllers.Logout)
	}
}
