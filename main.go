package main

import (
	"log"
	"routy/handlers"
	"routy/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	utils.InitRedis()
	router := gin.Default()
	// sample output for Rendering
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Routy API is live!"})
	})

	router.GET("/route", func(c *gin.Context) {
		from := c.Query("from")
		to := c.Query("to")
		categories := c.QueryArray("category") // hem tek hem çoklu destek

		if from == "" || to == "" {
			c.JSON(400, gin.H{"error": "from and to are required"})
			return
		}

		res, err := handlers.GetRouteData(from, to, categories)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, res)
	})
	router.POST("/route", handlers.HandleRoute)
	router.POST("/route_with_current", handlers.HandleRouteWithCurrent)

	log.Println("Server 8080 portunda başlatıldı...")

	err := router.Run(":8080")
	if err != nil {
		log.Println("sunucuya bağlanılamadı", err)
	}
}
