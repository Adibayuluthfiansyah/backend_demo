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

	config.ConnectDatabase()

	// ============================\
	// AUTOMIGRATE
	// ============================
	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.Document{},
		&models.SecretToken{},
		&models.SuperiorOrder{},
		&models.DocumentStaff{},
		&models.Notification{}, // <-- JANGAN LUPA TAMBAHKAN INI
	); err != nil {
		log.Fatal("Gagal migrasi tabel:", err)
	}

	// ============================\
	// GLOBAL MIDDLEWARE
	// ============================
	r.Use(middleware.RateLimiter())
	r.Use(middleware.CORSMiddleware())

	// ============================\
	// BUAT SATU GRUP /api
	// ============================
	api := r.Group("/api")
	{
		// Rute yang tidak perlu Auth
		routes.LoginRoutes(api)  // Kirim "api"
		routes.LogoutRoutes(api) // Kirim "api"

		// Rute yang perlu Auth (dikelola di dalam fungsi rute masing-masing)
		routes.UserRoutes(api)          // Kirim "api"
		routes.DocumentRoutes(api)      // Kirim "api"
		routes.DocumentStaffRoutes(api) // Kirim "api"
		routes.SuperiorOrderRoutes(api) // Kirim "api"

		// Rute Notifikasi BARU Anda
		routes.NotificationRoutes(api) // Kirim "api"
	}

	// ============================\
	// RUN SERVER
	// ============================
	log.Println("✅ Server berjalan di port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
