package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GeoLocation represents the structure of the response from the Nominatim API
type GeoLocation struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// GetCoordinates converts a given place name to its latitude and longitude using Nominatim API
func GetCoordinates(place string) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Add("q", place)
	params.Add("format", "json")
	params.Add("limit", "1")

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	log.Println("Nominatim request URL:", reqURL) // Added log

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", "routy-api/1.0")

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("Nominatim request failed:", err)
		return 0, 0, err
	}
	defer resp.Body.Close()

	var results []GeoLocation
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		log.Println("JSON decoding failed:", err)
		return 0, 0, err
	}

	log.Printf("Nominatim result: %+v\n", results)

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("location not found: %s", place)
	}

	latF, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, err
	}

	lonF, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, err
	}

	return latF, lonF, nil
}
