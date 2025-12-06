package controllers

import (
	"k-cms/config"
	"k-cms/models"
	"k-cms/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSiteConfig はサイト設定を取得する。設定がない場合はデフォルト値で作成して返す。
func GetSiteConfig(c *gin.Context) {
	var siteConfig models.SiteConfig
	err := config.DB.First(&siteConfig).Error
	if err != nil {
		// レコードがない場合は初期作成
		siteConfig = models.SiteConfig{
			SiteTitle:       "My Blog",
			SiteDescription: "Welcome to my blog.",
		}
		config.DB.Create(&siteConfig)
	}
	c.JSON(http.StatusOK, siteConfig)
}

// UpdateSiteConfig はサイト設定を更新し、バックグラウンドビルドをトリガーする。
func UpdateSiteConfig(c *gin.Context) {
	var input models.SiteConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var siteConfig models.SiteConfig
	if err := config.DB.First(&siteConfig).Error; err != nil {
		// レコードがない場合は新規作成として扱う
		siteConfig = input
		if err := config.DB.Create(&siteConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create site config"})
			return
		}
	} else {
		// 既存レコードの更新
		// IDは上書きしないよう注意（Firstで取得しているのでIDはセットされているはず）
		// GORMのUpdatesはゼロ値を無視するので、Model(&siteConfig).Updates(input)だとboolのfalse等が更新されない可能性があるが
		// 入力＝全項目送信前提であれば Save や特定カラム指定更新が良い。
		// ここでは簡便のため、inputの内容をsiteConfigにコピーしてSaveする。
		
		siteConfig.SiteTitle = input.SiteTitle
		siteConfig.SiteDescription = input.SiteDescription
		siteConfig.GoogleAnalyticsId = input.GoogleAnalyticsId
		siteConfig.OgpImageUrl = input.OgpImageUrl
		siteConfig.TwitterCardType = input.TwitterCardType
		siteConfig.TwitterSite = input.TwitterSite
		siteConfig.RobotIndex = input.RobotIndex
		siteConfig.PublisherType = input.PublisherType
		siteConfig.PublisherLogoUrl = input.PublisherLogoUrl
		siteConfig.PublisherDescription = input.PublisherDescription
		siteConfig.SocialLinks = input.SocialLinks

		if err := config.DB.Save(&siteConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update site config"})
			return
		}
	}

	// 設定変更後はビルドをトリガー
	go utils.TriggerBuild("update_site_config")

	c.JSON(http.StatusOK, gin.H{"message": "Site config updated and build triggered", "data": siteConfig})
}
