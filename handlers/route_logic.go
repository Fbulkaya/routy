package handlers

import (
	"routy/models"
	"routy/utils"
)

// Ortak mantık: POST ve GET için aynı
func GetRouteData(start, end string, interests []string) (models.RouteResponse, error) {
	startLat, startLon, err := utils.GetCoordinates(start)
	if err != nil {
		return models.RouteResponse{}, err
	}

	endLat, endLon, err := utils.GetCoordinates(end)
	if err != nil {
		return models.RouteResponse{}, err
	}

	places, err := utils.GetPlacesFromOverpassParallel(startLat, startLon, interests)
	if err != nil {
		places = []models.Place{}
	}

	if len(places) > 15 {
		places = places[:15]
	}

	return models.RouteResponse{
		Start:      start,
		End:        end,
		Stops:      places,
		StartCoord: []float64{startLat, startLon},
		EndCoord:   []float64{endLat, endLon},
	}, nil
}
