package routes

import (
	"dinsos_kuburaya/controllers"

	"github.com/gin-gonic/gin"
)

func LoginRoutes(r *gin.RouterGroup) {

	{
		r.POST("/login", controllers.Login)
	}
}
