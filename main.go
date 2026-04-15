package main

import (
	"k-cms/config"
	"k-cms/models"
	"k-cms/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	if os.Getenv("JWT_SECRET") == "" {
		panic("JWT_SECRET environment variable is not set. Please set it for security.")
	}

	config.ConnectDB()

	if err := config.DB.AutoMigrate(&models.User{}); err != nil {
		panic("Failed to migrate database.")
	}

	if err := config.DB.AutoMigrate(&models.Article{}); err != nil {
		panic("Failed to migrate database.")
	}

	if err := config.DB.AutoMigrate(&models.Image{}); err != nil {
		panic("Failed to migrate database.")
	}

	if err := config.DB.AutoMigrate(&models.SiteConfig{}); err != nil {
		panic("Failed to migrate database.")
	}

	if err := models.MigrateLike(config.DB); err != nil {
		panic("Failed to migrate like table.")
	}

	if err := models.MigratePageView(config.DB); err != nil {
		panic("Failed to migrate page_view table.")
	}

	router := gin.Default()
	routes.SetupRoutes(router)

	router.Run()
}
