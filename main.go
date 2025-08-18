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

	router.POST("/route", handlers.HandleRoute)
	router.POST("/route_with_current", handlers.HandleRouteWithCurrent)

	log.Println("Server 8080 portunda başlatıldı...")

	err := router.Run(":8080")
	if err != nil {
		log.Println("sunucuya bağlanılamadı", err)
	}
}
