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

	router := gin.Default()
	routes.SetupRoutes(router)

	router.Run()
}
