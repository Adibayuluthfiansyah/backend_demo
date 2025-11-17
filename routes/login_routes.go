package routes

import (
	"dinsos_kuburaya/controllers"

	"github.com/gin-gonic/gin"
)

// LoginRoutes mendaftarkan endpoint untuk login
func LoginRoutes(r *gin.RouterGroup) {

	{
		r.POST("/login", controllers.Login)
	}
}
