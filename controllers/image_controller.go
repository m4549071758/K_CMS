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

	var webpFileName string
	var webpFilePath string

	if fileExtension == ".webp" {
		webpFileName = id.String() + fileExtension
		webpFilePath = filepath.Join(uploadDir, webpFileName)
		if err := c.SaveUploadedFile(file, webpFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
			return
		}
	} else {
		// Webp変換は一時ファイルで処理する
		tempFileName := id.String() + fileExtension
		tempFilePath := filepath.Join(uploadDir, tempFileName)
		if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
			return
		}

		webpFileName = id.String() + ".webp"
		webpFilePath = filepath.Join(uploadDir, webpFileName)

		if err := convertToWebP(tempFilePath, webpFilePath, 80); err != nil {
			os.Remove(tempFilePath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to WebP"})
			return
		}

		os.Remove(tempFilePath)
	}

	image := models.Image{
		ID:       id,
		FileName: webpFileName,
	}

	if err := config.DB.Create(&image).Error; err != nil {
		os.Remove(webpFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

	fileURL := "https://www.katori.dev/api/images/" + webpFileName

	c.JSON(http.StatusCreated, ImageResponse{
		FileName: webpFileName,
		FileURL:  fileURL,
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

func GetImages(c *gin.Context) {
	var images []models.Image
	if err := config.DB.Order("updated_at desc").Find(&images).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch images"})
		return
	}

	var response []ImageResponse
	for _, img := range images {
		response = append(response, ImageResponse{
			FileName: img.FileName,
			FileURL:  "https://www.katori.dev/api/images/" + img.FileName,
		})
	}

	c.JSON(http.StatusOK, response)
}

func DeleteImage(c *gin.Context) {
	id := c.Param("id")
	var image models.Image

	// IDまたはファイル名で検索
	if err := config.DB.Where("id = ? OR file_name = ?", id, id).First(&image).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// 実ファイルの削除
	filePath := filepath.Join("images", image.FileName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		// ファイルが存在しない場合はDB削除のみ進めるため、エラーログだけ吐いて続行もありうるが
		// ここではエラーとして返す（または警告のみにするか要検討。今回はエラーログ出しつつ続行する方針で）
		// log.Printf("Failed to delete file: %v", err)
	}

	// DBレコードの削除
	if err := config.DB.Delete(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}
