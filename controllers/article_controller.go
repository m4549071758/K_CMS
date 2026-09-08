package controllers

import (
	"errors"
	"k-cms/config"
	"k-cms/middlewares"
	"k-cms/models"
	"k-cms/utils"

	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleInput struct {
	Title          string   `json:"title" binding:"required"`
	Excerpt        string   `json:"excerpt" binding:"required"`
	SeoTitle       string   `json:"seo_title" binding:"omitempty,max=255"`
	SeoDescription string   `json:"seo_description" binding:"omitempty,max=500"`
	PrimaryKeyword string   `json:"primary_keyword" binding:"omitempty,max=255"`
	CoverImageURL  string   `json:"cover_image" binding:"required"`
	OgImageURL     string   `json:"og_image" binding:"required"`
	Tags           []string `json:"tags" binding:"required"`
	Datetime       string   `json:"datetime" binding:"required"`
	Content        string   `json:"content" binding:"required"`
}

type ArticlesResponse struct {
	ArticleID string `json:"article_id"`
	Title     string `json:"title"`
	Excerpt   string `json:"excerpt"`
	LikeCount int    `json:"like_count"`
}

type ArticleResponse struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Excerpt        string   `json:"excerpt"`
	SeoTitle       string   `json:"seo_title"`
	SeoDescription string   `json:"seo_description"`
	CoverImageURL  string   `json:"cover_image"`
	OgImageURL     string   `json:"og_image"`
	Tags           []string `json:"tags"`
	Datetime       string   `json:"datetime"`
	Content        string   `json:"content"`
	LikeCount      int      `json:"like_count"`
}

type ArticleEditorResponse struct {
	ArticleResponse
	PrimaryKeyword string `json:"primary_keyword"`
}

func newArticleResponse(article models.Article) ArticleResponse {
	return ArticleResponse{
		ID:             article.ID.String(),
		Title:          article.Title,
		Excerpt:        article.Excerpt,
		SeoTitle:       article.SeoTitle,
		SeoDescription: article.SeoDescription,
		CoverImageURL:  article.CoverImageURL,
		OgImageURL:     article.OgImageURL,
		Tags:           article.Tags,
		Datetime:       article.Datetime,
		Content:        article.Content,
		LikeCount:      article.LikeCount,
	}
}

func GetArticles(c *gin.Context) {
	response := make([]ArticlesResponse, 0)

	log.Println("GetArticles: Starting to fetch articles")

	// 改善: Select で必要なカラムだけに絞り、Scan で直接 response に入れる
	// これにより本文 (Content) などの重いデータを読み込まない
	if err := config.DB.Model(&models.Article{}).
		Select("id as article_id, title, excerpt, like_count").
		Order("datetime desc"). // 日付順のソートを追加
		Scan(&response).Error; err != nil {
		log.Printf("GetArticles: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}

	log.Printf("GetArticles: Returning %d articles", len(response))
	c.JSON(http.StatusOK, response)
}

func GetArticle(c *gin.Context) {
	var article models.Article
	// パスパラメータから記事idを取得
	id := c.Param("id")

	if err := config.DB.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		log.Printf("GetArticle: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
		return
	}

	response := newArticleResponse(article)

	c.JSON(http.StatusOK, response)
}

func GetArticleForEditor(c *gin.Context) {
	var article models.Article
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		log.Printf("GetArticleForEditor: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch article"})
		return
	}

	userUUID, err := middlewares.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if userUUID != article.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to edit this article"})
		return
	}

	c.JSON(http.StatusOK, ArticleEditorResponse{
		ArticleResponse: newArticleResponse(article),
		PrimaryKeyword:  article.PrimaryKeyword,
	})
}

func AddArticle(c *gin.Context) {
	var input ArticleInput

	// 登録するときはいきなりCreateに突っ込むのではなく、ShouldBindJSONでバリデーションをかけてからcreate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := middlewares.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	article := models.Article{
		Title:          input.Title,
		Excerpt:        input.Excerpt,
		SeoTitle:       input.SeoTitle,
		SeoDescription: input.SeoDescription,
		PrimaryKeyword: input.PrimaryKeyword,
		CoverImageURL:  input.CoverImageURL,
		OgImageURL:     input.OgImageURL,
		Tags:           models.StringArray(input.Tags), // 直接StringArrayに変換
		Datetime:       input.Datetime,
		Content:        input.Content,
		UserID:         userUUID,
	}

	if err := config.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
		return
	}

	// Trigger frontend build for article creation
	utils.TriggerBuild("create", article.ID.String())

	config.DB.Preload("User").First(&article, article.ID)
	c.JSON(http.StatusCreated, article)
}

func UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article
	if err := config.DB.Where("id = ?", id).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	userUUID, err := middlewares.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if userUUID != article.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not allowed to update this article"})
		return
	}

	var input ArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article.Title = input.Title
	article.Excerpt = input.Excerpt
	article.SeoTitle = input.SeoTitle
	article.SeoDescription = input.SeoDescription
	article.PrimaryKeyword = input.PrimaryKeyword
	article.CoverImageURL = input.CoverImageURL
	article.OgImageURL = input.OgImageURL
	article.Tags = models.StringArray(input.Tags)
	article.Datetime = input.Datetime
	article.Content = input.Content

	var t time.Time

	t, err = time.Parse(time.RFC3339, input.Datetime)
	if err != nil {

		t, err = time.Parse("2006-01-02", input.Datetime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無効な日付形式です"})
			return
		}
	}

	article.Datetime = t.Format("2006-01-02")

	if err := config.DB.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update article"})
		return
	}

	// Trigger frontend build for article update
	utils.TriggerBuild("update", article.ID.String())

	config.DB.Preload("User").First(&article, article.ID)
	c.JSON(http.StatusOK, article)
}

func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article
	if err := config.DB.Where("id = ?", id).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	userUUID, err := middlewares.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if userUUID != article.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not allowed to delete this article"})
		return
	}

	if err := config.DB.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article"})
		return
	}

	// Trigger frontend build for article deletion
	utils.TriggerBuild("delete", article.ID.String())

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})
}
