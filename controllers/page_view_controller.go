package controllers

import (
	"k-cms/config"
	"k-cms/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type PageViewRequest struct {
	ArticleID   string `json:"article_id" binding:"required"`
	Fingerprint string `json:"fingerprint" binding:"required"`
}

type PageViewResponse struct {
	ArticleID string `json:"article_id"`
	ViewCount int64  `json:"view_count"`
	Message   string `json:"message"`
}

// RecordPageView は1記事・1fingerprint・1日で1回だけカウントを記録する。
// 既存レコードがある場合は何もせず現在のカウントを返す（冪等）。
func RecordPageView(c *gin.Context) {
	var input PageViewRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// フィンガープリントバリデーション
	if len(input.Fingerprint) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fingerprint"})
		return
	}

	// 記事の存在確認
	var article models.Article
	if err := config.DB.Where("id = ?", input.ArticleID).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// 今日の日付（時刻なしのDATE）
	today := time.Now().Truncate(24 * time.Hour)
	clientIP := c.ClientIP()

	// 同一記事・同一fingerprint・同一日のレコードを検索
	var existing models.PageView
	err := config.DB.Where(
		"article_id = ? AND fingerprint = ? AND visited_date = ?",
		input.ArticleID, input.Fingerprint, today,
	).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 初回アクセス → 記録する
		articleUUID, _ := uuid.FromString(input.ArticleID)
		newView := models.PageView{
			ArticleID:   articleUUID,
			Fingerprint: input.Fingerprint,
			IPAddress:   clientIP,
			VisitedDate: today,
		}
		if err := config.DB.Create(&newView).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record page view"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	// else: 既存レコードあり → スキップ（冪等）

	// 総PV数を返す
	var count int64
	config.DB.Model(&models.PageView{}).Where("article_id = ?", input.ArticleID).Count(&count)

	c.JSON(http.StatusOK, PageViewResponse{
		ArticleID: input.ArticleID,
		ViewCount: count,
		Message:   "OK",
	})
}

// GetPageViewCount は記事の総PV数を返す。
func GetPageViewCount(c *gin.Context) {
	articleID := c.Param("id")

	// 記事の存在確認
	var article models.Article
	if err := config.DB.Where("id = ?", articleID).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	var count int64
	config.DB.Model(&models.PageView{}).Where("article_id = ?", articleID).Count(&count)

	c.JSON(http.StatusOK, PageViewResponse{
		ArticleID: articleID,
		ViewCount: count,
		Message:   "OK",
	})
}
