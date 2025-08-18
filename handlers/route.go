package handlers

import (
	"log"
	"net/http"

	"routy/models"
	"routy/utils"

	"github.com/gin-gonic/gin"
)

// HandleRoute handles the standard route request without current location
func HandleRoute(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Panic occurred:", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	var req models.RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get coordinates for start and end addresses
	startLat, startLon, err := utils.GetCoordinates(req.Start)
	if err != nil {
		log.Println("Failed to get start coordinates:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start location"})
		return
	}

	endLat, endLon, err := utils.GetCoordinates(req.End)
	if err != nil {
		log.Println("Failed to get end coordinates:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end location"})
		return
	}

	// Fetch relevant places along the route
	// Fetch relevant places along the route
	places, err := utils.GetPlacesFromOverpassParallel(startLat, startLon, req.Interests)
	if err != nil {
		log.Println("Failed to get places:", err)
		places = []models.Place{}
	}

	// Limit result to 15 places
	if len(places) > 15 {
		places = places[:15]
	}

	res := models.RouteResponse{
		Start:      req.Start,
		End:        req.End,
		Stops:      places,
		StartCoord: []float64{startLat, startLon},
		EndCoord:   []float64{endLat, endLon},
	}

	log.Printf("Number of places returned: %d\n", len(places))
	c.JSON(http.StatusOK, res)
}

// HandleRouteWithCurrent handles route request with user's current location
func HandleRouteWithCurrent(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Panic occurred:", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	var req models.RouteRequestWithCurrent
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Convert start and end locations to coordinates
	startLat, startLon, err := utils.GetCoordinates(req.Start)
	if err != nil {
		log.Println("Failed to get start coordinates:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start location"})
		return
	}

	endLat, endLon, err := utils.GetCoordinates(req.End)
	if err != nil {
		log.Println("Failed to get end coordinates:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end location"})
		return
	}

	// Get places considering the current location as a context
	places, err := utils.GetPlacesBetweenWithCurrentLocation(startLat, startLon, endLat, endLon, req.Current.Lat, req.Current.Lon, req.Interests)
	if err != nil {
		log.Println("Failed to get places:", err)
		places = []models.Place{}
	}

	// Limit to max 15 results
	if len(places) > 15 {
		places = places[:15]
	}

	res := models.RouteResponse{
		Start:      req.Start,
		End:        req.End,
		Stops:      places,
		StartCoord: []float64{startLat, startLon},
		EndCoord:   []float64{endLat, endLon},
	}

	log.Printf("Number of places returned: %d\n", len(places))
	c.JSON(http.StatusOK, res)
}
