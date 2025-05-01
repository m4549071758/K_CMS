package controllers

import (
	"k-cms/config"
	"k-cms/middlewares"
	"k-cms/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func GetUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User

	// プレースホルダを使用してパラメータ化クエリを実行
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
func CreateUser(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindBodyWithJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	Register(c)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetString("user_id")

	if currentUserID != id {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not allowed to update this user"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	type UpdateUserInput struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.Username != "" {
		updates["username"] = input.Username
	}
	if input.Email != "" {
		updates["email"] = input.Email
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetString("user_id")

	if currentUserID != id {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not allowed to delete this user"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// パスワード変更用の入力構造体
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword はユーザーのパスワードを変更する関数
func ChangePassword(c *gin.Context) {
	// JWTトークンからユーザーを取得
	user, err := middlewares.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "認証されていません"})
		return
	}

	// リクエストからパスワード情報を取得
	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 現在のパスワードを検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "現在のパスワードが正しくありません"})
		return
	}

	// 新しいパスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "パスワードのハッシュ化に失敗しました"})
		return
	}

	// パスワードを更新
	if err := config.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "パスワードの更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "パスワードが正常に変更されました"})
}
