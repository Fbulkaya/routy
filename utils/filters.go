package utils

import (
	"log"
	"routy/models"
)

// FilterStopsByDistance filters the given places based on proximity to route coordinates.
// Only places within maxDistanceMeters from any point on the route are included.
func FilterStopsByDistance(routeCoords [][]float64, places []models.Place, maxDistanceMeters float64) []models.Place {
	filtered := make([]models.Place, 0)

	for _, place := range places {
		for _, coord := range routeCoords {
			// Calculate distance between place and current route coordinate
			dist := haversineDistance(place.Latitude, place.Longitude, coord[1], coord[0])

			// If within allowed distance, include it and move to next place
			if dist <= maxDistanceMeters {
				log.Printf("Matched place: %s → Distance: %.2f m", place.Name, dist)
				filtered = append(filtered, place)
				break
			}
		}
	}

	log.Printf("Total number of matched places: %d", len(filtered))
	return filtered
}
