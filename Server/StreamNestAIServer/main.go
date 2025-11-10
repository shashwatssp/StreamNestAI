package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/database"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/routes"
)

func main() {
	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "StreamNest")
	})

	var client *mongo.Client = database.Connect()

	routes.SetupUnProtectedRoutes(router, client)
	routes.SetupProtectedRoutes(router, client)

	if err := router.Run(": 8080"); err != nil {
		fmt.Println("ERROR ", err)
	}
}
