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

	// 記事の存在確認（IDだけ確認）
	var article models.Article
	if err := config.DB.Select("id, like_count").Where("id = ?", input.ArticleID).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// フィンガープリントバリデーション
	if len(input.Fingerprint) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fingerprint"})
		return
	}

	clientIP := c.ClientIP()
	articleUUID, _ := uuid.FromString(input.ArticleID)

	var response LikeResponse

	// トランザクション開始
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var existingLike models.Like
		err := tx.Unscoped().Where("article_id = ? AND fingerprint = ?", input.ArticleID, input.Fingerprint).First(&existingLike).Error

		if err == gorm.ErrRecordNotFound {
			// 新規作成
			newLike := models.Like{
				ArticleID:   articleUUID,
				Fingerprint: input.Fingerprint,
				IPAddress:   clientIP,
			}
			if err := tx.Create(&newLike).Error; err != nil {
				return err
			}
			// カウントアップ
			if err := tx.Model(&article).Update("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
				return err
			}
			response = LikeResponse{IsLiked: true, Message: "Like added"}
		} else if err != nil {
			return err
		} else {
			if existingLike.DeletedAt.Valid {
				// 復元
				if err := tx.Unscoped().Model(&existingLike).Update("deleted_at", nil).Error; err != nil {
					return err
				}
				if err := tx.Model(&article).Update("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
					return err
				}
				response = LikeResponse{IsLiked: true, Message: "Like restored"}
			} else {
				// 削除
				if err := tx.Delete(&existingLike).Error; err != nil {
					return err
				}
				if err := tx.Model(&article).Update("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
					return err
				}
				response = LikeResponse{IsLiked: false, Message: "Like removed"}
			}
		}

		// 最新のカウントを取得（Selectで絞り込み）
		if err := tx.Model(&article).Select("like_count").First(&article).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	response.ArticleID = input.ArticleID
	response.LikeCount = article.LikeCount
	c.JSON(http.StatusOK, response)
}

// 記事のいいね状態を取得
func GetLikeStatus(c *gin.Context) {
	articleID := c.Param("id")
	fingerprint := c.Query("fingerprint")

	if fingerprint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fingerprint is required"})
		return
	}

	// 記事の存在確認（Selectで絞り込み）
	var article models.Article
	if err := config.DB.Select("id, like_count").Where("id = ?", articleID).First(&article).Error; err != nil {
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
