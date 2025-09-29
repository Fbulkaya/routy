package handlers

import (
	"log"
	"net/http"
	"routy/models"
	"routy/utils"

	"github.com/gin-gonic/gin"
)

// HandleRouteGET handles GET requests with query params (?from=...&to=...&category=...)
func HandleRouteGET(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	category := c.Query("category")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to parameters are required"})
		return
	}

	// Adresleri koordinatlara çevir
	startLat, startLon, err := utils.GetCoordinates(from)
	if err != nil {
		log.Println("Failed to get start coordinates:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start location"})
		return
	}

	endLat, endLon, err := utils.GetCoordinates(to)
	if err != nil {
		log.Println("Failed to get end coordinates:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end location"})
		return
	}

	// Overpass'ten kategoriye göre mekanları al
	places, err := utils.GetPlacesFromOverpassParallel(startLat, startLon, []string{category})
	if err != nil {
		log.Println("Failed to get places:", err)
		places = []models.Place{}
	}

	// En fazla 15 sonuç
	if len(places) > 15 {
		places = places[:15]
	}

	res := models.RouteResponse{
		Start:      from,
		End:        to,
		Stops:      places,
		StartCoord: []float64{startLat, startLon},
		EndCoord:   []float64{endLat, endLon},
	}

	c.JSON(http.StatusOK, res)
}
