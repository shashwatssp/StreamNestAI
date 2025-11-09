package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"

	controller "github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/controllers"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/database"
)

func main() {
	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "StreamNest")
	})

	var client *mongo.Client = database.Connect()

	router.GET("/movies", controller.GetMovies(client))

	router.GET("/movie/:imdb_id", controller.GetMovie(client))

	router.POST("/addmovie", controller.AddMovie(client))

	router.POST("/register", controller.RegisterUser(client))

	if err := router.Run(": 8080"); err != nil {
		fmt.Println("ERROR ", err)
	}
}
