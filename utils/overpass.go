package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"routy/models"
)

// Paralel Overpass çağrısı yapan fonksiyon
func GetPlacesFromOverpassParallel(lat, lon float64, interests []string) ([]models.Place, error) {
	var wg sync.WaitGroup
	placeChan := make(chan []models.Place, len(interests))
	errChan := make(chan error, len(interests))

	client := &http.Client{Timeout: 15 * time.Second}

	for _, interest := range interests {
		wg.Add(1)
		go func(interest string) {
			defer wg.Done()

			query := fmt.Sprintf(`[out:json];node[amenity=%s](around:300,%.6f,%.6f);out;`, interest, lat, lon)
			url := "https://overpass-api.de/api/interpreter?data=" + query

			resp, err := client.Get(url)
			if err != nil {
				errChan <- err
				return
			}
			defer resp.Body.Close()

			var result struct {
				Elements []struct {
					Lat  float64           `json:"lat"`
					Lon  float64           `json:"lon"`
					Tags map[string]string `json:"tags"`
				} `json:"elements"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				errChan <- err
				return
			}

			var places []models.Place
			for _, el := range result.Elements {
				places = append(places, models.Place{
					Name:      el.Tags["name"],
					Latitude:  el.Lat,
					Longitude: el.Lon,
					Category:  interest,
				})
			}

			placeChan <- places
		}(interest)
	}

	wg.Wait()
	close(placeChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan
	}

	var allPlaces []models.Place
	for p := range placeChan {
		allPlaces = append(allPlaces, p...)
	}

	return allPlaces, nil
}
