package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/controllers"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	// Create a protected route group instead of applying middleware globally
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleWare())
	{
		protected.GET("/movie/:imdb_id", controller.GetMovie(client))
		protected.POST("/addmovie", controller.AddMovie(client))
		protected.GET("/recommendedmovies", controller.GetRecommendedMovies(client))
		protected.PATCH("/updatereview/:imdb_id", controller.AdminReviewUpdate(client))
	}
}
