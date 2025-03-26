package controllers

import (
	"k-cms/config"
	"k-cms/models"
	"net/http"

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

func GetArticles(c *gin.Context) {
	var articles []models.Article
	if err := config.DB.Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}

	c.JSON(http.StatusOK, articles)
}

func GetArticle(c *gin.Context) {
	var Article models.Article
	// パスパラメータから記事idを取得
	id := c.Param("id")

	if err := config.DB.Where("id = ?", id).First(&Article).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, Article)
}

func AddArticle(c *gin.Context) {
	var input ArticleInput

	// 登録するときはいきなりCreateに突っ込むのではなく、ShouldBindJSONでバリデーションをかけてからcreate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	}

	if err := config.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
		return
	}

	c.JSON(http.StatusCreated, article)
}
