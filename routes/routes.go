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

		protected.POST("/articles/add", controllers.AddArticle)
		protected.PUT("/articles/:id", controllers.UpdateArticle)
		protected.DELETE("/articles/:id", controllers.DeleteArticle)

		protected.GET("/is_Auth", controllers.IsAuthenticated)
	}
}
