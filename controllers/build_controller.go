package controllers

import (
	"k-cms/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBuildStatus は現在のビルド状態を返します
func GetBuildStatus(c *gin.Context) {
	status := utils.GetBuildStatus()
	c.JSON(http.StatusOK, status)
}
