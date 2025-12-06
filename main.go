package main

import (
	"k-cms/config"
	"k-cms/models"
	"k-cms/routes"

	"github.com/gin-gonic/gin"
)

func main() {
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

	router := gin.Default()
	routes.SetupRoutes(router)

	router.Run()
}
