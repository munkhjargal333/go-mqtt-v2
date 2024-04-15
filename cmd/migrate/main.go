package main

import (
	"fmt"
	"log"
	"mqtt/database"
	"mqtt/models"

	"github.com/joho/godotenv"
)

// var DB *gorm.DB

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	database.MustConnect()
	db := database.DB

	// Migrate the Content0230 model
	// if err := DB.AutoMigrate(&models.Content0230{}); err != nil {
	// 	log.Fatal("Error migrating Content0230 model:", err)
	// 	return
	// }
	// if err := DB.AutoMigrate(&models.Content0200{}); err != nil {
	// 	log.Fatal("Error migrating Content0230 model:", err)
	// }

	// if err := DB.AutoMigrate(&models.Content0DE0{}); err != nil {
	// 	log.Fatal("Error migrating Content0230 model:", err)
	// }

	if err := db.AutoMigrate(&models.Content0800{}); err != nil {
		log.Fatal("Error migrating Content0230 model:", err)
	}

	// if err := DB.AutoMigrate(&models.Content0202{}); err != nil {
	// 	log.Fatal("Error migrating Content0230 model:", err)
	// }

	// if err := db.AutoMigrate(&models.FailedLog{}); err != nil {
	// 	log.Fatal("Error migrating FailedLog model:", err)
	// }
}
