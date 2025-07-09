package middlewares

import (
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

func CSRFMiddleware() gin.HandlerFunc {
	return csrf.Middleware(csrf.Options{
		Secret: "csrf-secret-key-change-in-production", // 本番環境では環境変数から読み込み
		ErrorFunc: func(c *gin.Context) {
			c.JSON(403, gin.H{"error": "CSRF token mismatch"})
			c.Abort()
		},
	})
}
