package middlewares

import (
	"errors"
	"k-cms/config"
	"k-cms/models"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt"
)

// トークン認証用のミドルウェア
func AuthMiddleware() gin.HandlerFunc {

	// Authorizationヘッダーのチェック
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required."})
			c.Abort()
			return
		}

		// トークン文字列を分割、 Bearer <token>の形式かチェックし、配列に挿入
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer <token>."})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userIDStr := claims["user_id"].(string)
			userID, err := uuid.FromString(userIDStr)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
				c.Abort()
				return
			}

			var user models.User
			if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found."})
				c.Abort()
				return
			}

			// ここでリクエスト処理前にContextに対して"user_id"と"user"をセット
			// リクエスト処理中にこれらの値を取得できるようになる
			c.Set("user_id", userID)
			c.Set("user", user)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token."})
			c.Abort()
		}

	}
}

func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	// コンテキストからユーザーIDを取得
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user not authenticated")
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user_id is not a valid UUID")
	}

	return userUUID, nil
}

func GetUserFromContext(c *gin.Context) (models.User, error) {
	// コンテキストからユーザーを取得
	user, exists := c.Get("user")
	if !exists {
		return models.User{}, errors.New("user not authenticated")
	}

	// ユーザーがUser型かチェック
	userModel, ok := user.(models.User)
	if !ok {
		return models.User{}, errors.New("user is not a valid User model")
	}

	return userModel, nil
}
