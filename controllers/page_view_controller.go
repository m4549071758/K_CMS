package controllers

import (
	"errors"
	"k-cms/config"
	"k-cms/models"
	"net/http"
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

// isDuplicateKeyError は MySQL の Duplicate entry エラー（1062）かどうかを判定する。
func isDuplicateKeyError(err error) bool {
	var mysqlErr *gomysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

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

	// 今日の日付をローカルタイムゾーン基準の DATE 文字列で取得する。
	// time.Truncate(24h) はUTC基準のため JST(UTC+9) では MySQL の DATE 型との比較がずれる。
	todayStr := time.Now().Local().Format("2006-01-02")
	clientIP := c.ClientIP()

	// 同一記事・同一fingerprint・同一日のレコードを検索
	var existing models.PageView
	err := config.DB.Where(
		"article_id = ? AND fingerprint = ? AND visited_date = ?",
		input.ArticleID, input.Fingerprint, todayStr,
	).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 初回アクセス → 記録する
		articleUUID, _ := uuid.FromString(input.ArticleID)
		newView := models.PageView{
			ArticleID:   articleUUID,
			Fingerprint: input.Fingerprint,
			IPAddress:   clientIP,
			VisitedDate: time.Now().Local(),
		}
		if createErr := config.DB.Create(&newView).Error; createErr != nil {
			// UNIQUE制約違反（タイムゾーンズレや競合リクエスト等）はスキップ扱いとし
			// カウントを返す。それ以外のDBエラーは 500 を返す。
			if !isDuplicateKeyError(createErr) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record page view"})
				return
			}
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	// else: 既存レコードあり → スキップ（冪等）

	// 総PV数を返す（INSERT成功・スキップ・既存どのケースでも必ずここに到達する）
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
