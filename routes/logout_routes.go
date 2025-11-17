package routes

import (
	"dinsos_kuburaya/controllers"

	"github.com/gin-gonic/gin"
)

func LogoutRoutes(r *gin.RouterGroup) {
	{
		r.POST("/logout", controllers.Logout)
	}
}
