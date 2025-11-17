package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"dinsos_kuburaya/config"
	"dinsos_kuburaya/middleware"
	"dinsos_kuburaya/models"
	"dinsos_kuburaya/routes"
)

func main() {
	r := gin.Default()
	r.MaxMultipartMemory = 100 << 20

	// ============================
	// CONNECT DATABASE FIRST
	// ============================
	config.ConnectDatabase()

	// ============================
	// AUTOMIGRATE
	// ============================
	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.Document{},
		&models.SecretToken{},
		&models.SuperiorOrder{},
		&models.DocumentStaff{},
	); err != nil {
		log.Fatal("Gagal migrasi tabel:", err)
	}

	// ============================
	// GLOBAL MIDDLEWARE
	// ============================
	r.Use(middleware.RateLimiter())
	r.Use(middleware.CORSMiddleware())

	// ============================
	// ROUTES
	// ============================
	routes.LoginRoutes(r) // biasanya login tidak pakai AuthMiddleware
	routes.LogoutRoutes(r)
	routes.UserRoutes(r)
	routes.DocumentRoutes(r)
	routes.DocumentStaffRoutes(r)
	routes.SuperiorOrderRoutes(r)

	// ============================
	// RUN SERVER
	// ============================
	r.Run(":8080")
}
