package routes

import (
	"k-cms/controllers"
	"k-cms/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	r.Use(middlewares.CORSMiddleware())

	public := r.Group("/api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.GET("/articles", controllers.GetArticles)
		public.GET("/articles/:id", controllers.GetArticle)
		public.GET("/images/:filename", controllers.GetImage)
		public.GET("/like-status/:id", controllers.GetLikeStatus)

		// いいね機能をpublicに移動（fingerprintで同一性を判定）
		public.POST("/articles/like", controllers.ToggleLike)
		
		public.GET("/users/:id", controllers.GetUser) // プロフィール表示用
		public.GET("/site-config", controllers.GetSiteConfig) // サイト設定用
	}

	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware())
	{
		// GET("/エンドポイント:XXX")でパスパラメータが取れる

		protected.GET("/users", controllers.GetUsers)
		// protected.GET("/users/:id", controllers.GetUser) // Publicに移動済み
		protected.POST("/users", controllers.CreateUser)
		protected.PUT("/users/:id", controllers.UpdateUser)
		protected.DELETE("/users/:id", controllers.DeleteUser)
		protected.POST("/change-password", controllers.ChangePassword)

		protected.POST("/articles/add", controllers.AddArticle)
		protected.PUT("/articles/:id", controllers.UpdateArticle)
		protected.DELETE("/articles/:id", controllers.DeleteArticle)

		protected.GET("/is_Auth", controllers.IsAuthenticated)

		protected.POST("/images/upload", controllers.UploadImage)
		protected.GET("/images", controllers.GetImages)
		protected.DELETE("/images/:id", controllers.DeleteImage)
		protected.GET("/build-status", controllers.GetBuildStatus)

		// protected.GET("/site-config", controllers.GetSiteConfig) // Publicに移動済み
		protected.PUT("/site-config", controllers.UpdateSiteConfig)
	}
}
