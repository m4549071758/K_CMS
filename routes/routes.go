package routes

import (
	"k-cms/controllers"
	"k-cms/middlewares"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

func SetupRoutes(r *gin.Engine) {
	// セッションストアを設定
	store := cookie.NewStore([]byte("session-secret-key-change-in-production"))
	r.Use(sessions.Sessions("csrf-session", store))

	r.Use(middlewares.CORSMiddleware())

	// CSRFトークン取得エンドポイント（CSRFミドルウェア適用）
	csrf_group := r.Group("/")
	csrf_group.Use(middlewares.CSRFMiddleware())
	{
		csrf_group.GET("/csrf-token", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"csrf_token": csrf.GetToken(c),
			})
		})
		// いいね機能はCSRF保護が必要
		csrf_group.POST("/api/articles/like", controllers.ToggleLike)
	}

	public := r.Group("/api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.GET("/articles", controllers.GetArticles)
		public.GET("/articles/:id", controllers.GetArticle)
		public.GET("/images/:filename", controllers.GetImage)
		public.GET("/articles/:id/like-status", controllers.GetLikeStatus)
	}

	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware())
	{
		// GET("/エンドポイント:XXX")でパスパラメータが取れる

		protected.GET("/users", controllers.GetUsers)
		protected.GET("/users/:id", controllers.GetUser)
		protected.POST("/users", controllers.CreateUser)
		protected.PUT("/users/:id", controllers.UpdateUser)
		protected.DELETE("/users/:id", controllers.DeleteUser)
		protected.POST("/change-password", controllers.ChangePassword)

		protected.POST("/articles/add", controllers.AddArticle)
		protected.PUT("/articles/:id", controllers.UpdateArticle)
		protected.DELETE("/articles/:id", controllers.DeleteArticle)

		protected.GET("/is_Auth", controllers.IsAuthenticated)

		protected.POST("/images/upload", controllers.UploadImage)
	}
}
