package controllers

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"k-cms/config"
	"k-cms/models"
	"net/http"
	"os"
	"path/filepath"

	"github.com/chai2010/webp"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

type ImageInput struct {
	ArticleID string `form:"article_id" binding:"required"`
}

type ImageResponse struct {
	FileName  string `json:"file_name"`
	ArticleID string `json:"article_id"`
	FileURL   string `json:"file_url"`
}

func UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload image"})
		return
	}

	var input ImageInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileExtension := filepath.Ext(file.Filename)
	id, err := uuid.NewV7()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate UUID"})
		return
	}

	uploadDir := "images"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// 一時ファイルで処理する
	tempFileName := id.String() + fileExtension
	tempFilePath := filepath.Join(uploadDir, tempFileName)
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}

	webpFileName := id.String() + ".webp"
	webpFilePath := filepath.Join(uploadDir, webpFileName)

	if err := convertToWebP(tempFilePath, webpFilePath, 80); err != nil {
		os.Remove(tempFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to WebP"})
		return
	}

	os.Remove(tempFilePath)

	image := models.Image{
		ID:        id,
		FileName:  webpFileName,
		ArticleID: uuid.FromStringOrNil(input.ArticleID),
	}

	if err := config.DB.Create(&image).Error; err != nil {
		os.Remove(webpFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

	fileURL := "https://www.katori.dev/api/images/" + webpFileName

	c.JSON(http.StatusCreated, ImageResponse{
		FileName:  webpFileName,
		ArticleID: input.ArticleID,
		FileURL:   fileURL,
	})
}

func convertToWebP(inputPath, outputPath string, quality float32) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	options := &webp.Options{
		Lossless: false,
		Quality:  quality,
	}
	if err := webp.Encode(output, img, options); err != nil {
		return err
	}

	return nil
}

func GetImage(c *gin.Context) {
	fileName := c.Param("filename")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File name is required"})
		return
	}

	filePath := filepath.Join("images", fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(filePath)
}
