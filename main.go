package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"sentul-golf-be/config"
	"sentul-golf-be/models"
	"sentul-golf-be/routes"
	"sentul-golf-be/utils"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to database
	config.ConnectDB()

	// Connect to Redis for caching
	config.ConnectRedis()

	// Create upload directories if they don't exist
	createUploadDirectories()

	// Auto migrate database schemas
	db := config.GetDB()
	if err := db.AutoMigrate(
		&models.User{},
		&models.News{},
		&models.Event{},
		&models.Hole{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create default admin user if not exists
	createDefaultAdmin()

	// Setup routes
	router := routes.SetupRoutes()

	// Start server
	port := config.GetEnv("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)
	
	log.Printf("Server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func createUploadDirectories() {
	directories := []string{
		"./uploads",
		"./uploads/news",
		"./uploads/events",
		"./uploads/holes",
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Warning: Failed to create directory %s: %v", dir, err)
		}
	}
}

func createDefaultAdmin() {
	db := config.GetDB()
	var count int64
	db.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)

	if count == 0 {
		hashedPassword, err := utils.HashPassword("admin123")
		if err != nil {
			log.Println("Failed to create default admin:", err)
			return
		}

		admin := models.User{
			Name:     "Admin",
			Email:    "admin@sentulgolf.com",
			Password: hashedPassword,
			Role:     models.RoleAdmin,
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Println("Failed to create default admin:", err)
			return
		}

		log.Println("Default admin created - Email: admin@sentulgolf.com, Password: admin123")
	}
}
