package controllers

import (
	"k-cms/config"
	"k-cms/middlewares"
	"k-cms/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticleInput struct {
	Title         string   `json:"title" binding:"required"`
	Excerpt       string   `json:"excerpt" binding:"required"`
	CoverImageURL string   `json:"cover_image" binding:"required"`
	OgImageURL    string   `json:"og_image" binding:"required"`
	Tags          []string `json:"tags" binding:"required"`
	Datetime      string   `json:"datetime" binding:"required"`
	Content       string   `json:"content" binding:"required"`
}

type ArticlesResponse struct {
	ArticleID string `json:"article_id"`
	Title     string `json:"title"`
	Excerpt   string `json:"excerpt"`
}

type ArticleResponse struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Excerpt       string   `json:"excerpt"`
	CoverImageURL string   `json:"cover_image"`
	OgImageURL    string   `json:"og_image"`
	Tags          []string `json:"tags"`
	Datetime      string   `json:"datetime"`
	Content       string   `json:"content"`
}

func GetArticles(c *gin.Context) {
	var articles []models.Article
	var response []ArticlesResponse
	if err := config.DB.Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}

	// 取得した記事をresponseに詰め替え
	for _, article := range articles {
		response = append(response, ArticlesResponse{
			ArticleID: article.ID.String(),
			Title:     article.Title,
			Excerpt:   article.Excerpt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func GetArticle(c *gin.Context) {
	var article models.Article
	var response ArticleResponse
	// パスパラメータから記事idを取得
	id := c.Param("id")

	if err := config.DB.Where("id = ?", id).First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	response = ArticleResponse{
		ID:            article.ID.String(),
		Title:         article.Title,
		Excerpt:       article.Excerpt,
		CoverImageURL: article.CoverImageURL,
		OgImageURL:    article.OgImageURL,
		Tags:          article.Tags,
		Datetime:      article.Datetime,
		Content:       article.Content,
	}

	c.JSON(http.StatusOK, response)
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
		Title:         input.Title,
		Excerpt:       input.Excerpt,
		CoverImageURL: input.CoverImageURL,
		OgImageURL:    input.OgImageURL,
		Tags:          models.StringArray(input.Tags), // 直接StringArrayに変換
		Datetime:      input.Datetime,
		Content:       input.Content,
		UserID:        userUUID,
	}

	if err := config.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})
}
