package controllers

import (
	"k-cms/config"
	"k-cms/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type LikeRequest struct {
	ArticleID   string `json:"article_id" binding:"required"`
	Fingerprint string `json:"fingerprint" binding:"required"`
}

type LikeResponse struct {
	ArticleID string `json:"article_id"`
	LikeCount int    `json:"like_count"`
	IsLiked   bool   `json:"is_liked"`
	Message   string `json:"message"`
}

// いいねを追加または削除
func ToggleLike(c *gin.Context) {
	var input LikeRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 記事の存在確認
	var article models.Article
	if err := config.DB.Where("id = ?", input.ArticleID).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// フィンガープリントバリデーション
	if len(input.Fingerprint) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fingerprint"})
		return
	}

	// IPアドレス取得
	clientIP := c.ClientIP()

	// 既存のいいねをチェック（論理削除されたものも含む）
	var existingLike models.Like
	err := config.DB.Unscoped().Where("article_id = ? AND fingerprint = ?", input.ArticleID, input.Fingerprint).First(&existingLike).Error

	if err == gorm.ErrRecordNotFound {
		// いいねが存在しない場合、新規作成
		articleUUID, _ := uuid.FromString(input.ArticleID)
		newLike := models.Like{
			ArticleID:   articleUUID,
			Fingerprint: input.Fingerprint,
			IPAddress:   clientIP,
		}

		if err := config.DB.Create(&newLike).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create like"})
			return
		}

		// 記事のいいね数を更新
		if err := config.DB.Model(&article).Update("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update like count"})
			return
		}

		// 更新された記事情報を取得
		config.DB.First(&article, article.ID)

		c.JSON(http.StatusOK, LikeResponse{
			ArticleID: input.ArticleID,
			LikeCount: article.LikeCount,
			IsLiked:   true,
			Message:   "Like added",
		})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// いいねが存在する場合
		if existingLike.DeletedAt.Valid {
			// 論理削除されている場合、復元
			if err := config.DB.Unscoped().Model(&existingLike).Update("deleted_at", nil).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore like"})
				return
			}

			// 記事のいいね数を更新
			if err := config.DB.Model(&article).Update("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update like count"})
				return
			}

			// 更新された記事情報を取得
			config.DB.First(&article, article.ID)

			c.JSON(http.StatusOK, LikeResponse{
				ArticleID: input.ArticleID,
				LikeCount: article.LikeCount,
				IsLiked:   true,
				Message:   "Like restored",
			})
		} else {
			// 論理削除されていない場合、削除
			if err := config.DB.Delete(&existingLike).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete like"})
				return
			}

			// 記事のいいね数を更新
			if err := config.DB.Model(&article).Update("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update like count"})
				return
			}

			// 更新された記事情報を取得
			config.DB.First(&article, article.ID)

			c.JSON(http.StatusOK, LikeResponse{
				ArticleID: input.ArticleID,
				LikeCount: article.LikeCount,
				IsLiked:   false,
				Message:   "Like removed",
			})
		}
	}
}

// 記事のいいね状態を取得
func GetLikeStatus(c *gin.Context) {
	articleID := c.Param("id")
	fingerprint := c.Query("fingerprint")

	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fingerprint is required"})
		return
	}

	// 記事の存在確認
	var article models.Article
	if err := config.DB.Where("id = ?", articleID).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// いいね状態をチェック
	var existingLike models.Like
	isLiked := config.DB.Where("article_id = ? AND fingerprint = ?", articleID, fingerprint).First(&existingLike).Error == nil

	c.JSON(http.StatusOK, LikeResponse{
		ArticleID: articleID,
		LikeCount: article.LikeCount,
		IsLiked:   isLiked,
		Message:   "Like status retrieved",
	})
}
