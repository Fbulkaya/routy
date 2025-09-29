package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"routy/models"
	"strings"
	"time"
)

var isDebug = os.Getenv("APP_ENV") != "production"

// Represents the route geometry structure returned by OSRM
type RouteGeometry struct {
	Routes []struct {
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

// Finds places between the start and end points, based on interests
func GetPlacesBetween(startLat, startLon, endLat, endLon float64, interests []string) ([]models.Place, error) {
	// Construct the OSRM route API URL
	osrmURL := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson", startLon, startLat, endLon, endLat)
	resp, err := http.Get(osrmURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var routeData RouteGeometry
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &routeData); err != nil {
		return nil, err
	}
	// If route could not be found
	if len(routeData.Routes) == 0 {
		return nil, fmt.Errorf("route not found")
	}

	coordinates := routeData.Routes[0].Geometry.Coordinates
	placeSet := make(map[string]bool)
	var allPlaces []models.Place
	// Limit number of Overpass queries to avoid overload
	maxQueries := 15
	step := len(coordinates) / maxQueries
	if step < 1 {
		step = 1
	}

	for i := 0; i < len(coordinates); i += step {
		lon := coordinates[i][0]
		lat := coordinates[i][1]

		places, err := GetPlacesFromOverpassParallel(lat, lon, interests)
		if err != nil {
			log.Println("Error:", err)
			continue
		}

		for _, p := range places {
			key := fmt.Sprintf("%s_%f_%f", p.Name, p.Latitude, p.Longitude)
			if !placeSet[key] {
				placeSet[key] = true
				allPlaces = append(allPlaces, p)
			}
		}

	}

	filteredPlaces := FilterStopsByDistance(coordinates, allPlaces, 3000)
	if len(filteredPlaces) > 15 {
		filteredPlaces = filteredPlaces[:15]
	}
	return filteredPlaces, nil
}

// Makes Overpass API call or returns from Redis cache if available
func GetPlacesFromOverpass(lat1, lon1, lat2, lon2 float64, interest string) ([]models.Place, error) {
	if isDebug {
		fmt.Printf("[DEBUG] GetPlacesFromOverpass called with lat1=%.6f, lon1=%.6f, lat2=%.6f, lon2=%.6f, interest=%s\n",
			lat1, lon1, lat2, lon2, interest)
	}

	tag := mapInterestToOverpassTag(interest)
	cacheKey := fmt.Sprintf("overpass:route:%s:%.4f,%.4f:%.4f,%.4f", interest, lat1, lon1, lat2, lon2)

	// Redis kontrol
	if cached, err := RedisClient.Get(Ctx, cacheKey).Result(); err == nil {
		fmt.Println("[DEBUG] Data retrieved from Redis.")
		var results []models.Place
		if json.Unmarshal([]byte(cached), &results) == nil {
			return results, nil
		}
	}
	if isDebug {
		fmt.Println("[DEBUG] No data in Redis, querying OSRM & Overpass API...")
	}
	// OSRM'den rota noktalarını al
	osrmURL := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%.6f,%.6f;%.6f,%.6f?overview=full&geometries=geojson", lon1, lat1, lon2, lat2)
	resp, err := http.Get(osrmURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var osrmData struct {
		Routes []struct {
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&osrmData); err != nil {
		return nil, err
	}

	if len(osrmData.Routes) == 0 {
		return nil, fmt.Errorf("no route found")
	}

	coords := osrmData.Routes[0].Geometry.Coordinates

	// Her 500 metre aralıkla nokta seç
	var selectedPoints [][2]float64
	lastPoint := coords[0]
	selectedPoints = append(selectedPoints, [2]float64{lastPoint[1], lastPoint[0]}) // lat, lon
	const intervalMeters = 500

	for _, pt := range coords {
		if HaversineDistance(lastPoint[1], lastPoint[0], pt[1], pt[0]) >= intervalMeters {
			selectedPoints = append(selectedPoints, [2]float64{pt[1], pt[0]})
			lastPoint = pt
		}
	}

	// Nokta sayısını sınırlayalım
	if len(selectedPoints) > 10 {
		selectedPoints = selectedPoints[:10]
	}
	fmt.Printf("[DEBUG] Selected %d points along the route for Overpass queries\n", len(selectedPoints))

	// Overpass sorguları
	var results []models.Place
	seen := make(map[int64]bool)

	for _, p := range selectedPoints {
		lat, lon := p[0], p[1]
		if isDebug {
			fmt.Printf("[DEBUG] Sending Overpass query for point (%.6f, %.6f) and tag=%s\n", lat, lon, tag)
		}
		query := fmt.Sprintf(`
			[out:json][timeout:25];
			(
			  node[%s](around:300,%.6f,%.6f);
			  way[%s](around:300,%.6f,%.6f);
			  relation[%s](around:300,%.6f,%.6f);
			);
			out center;
		`, tag, lat, lon, tag, lat, lon, tag, lat, lon)

		resp, err := http.Post("https://overpass-api.de/api/interpreter", "application/x-www-form-urlencoded",
			strings.NewReader("data="+url.QueryEscape(query)))
		if err != nil {
			fmt.Println("[WARN] Overpass error:", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var data struct {
			Elements []struct {
				Type   string            `json:"type"`
				Lat    float64           `json:"lat"`
				Lon    float64           `json:"lon"`
				ID     int64             `json:"id"`
				Tags   map[string]string `json:"tags"`
				Center struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"center"`
			} `json:"elements"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			fmt.Println("[WARN] JSON unmarshal error:", err)
			continue
		}
		if isDebug {
			fmt.Printf("[DEBUG] Overpass returned %d elements for point (%.6f, %.6f)\n", len(data.Elements), lat, lon)
		}
		for _, el := range data.Elements {
			if seen[el.ID] {
				continue
			}
			seen[el.ID] = true

			plLat, plLon := el.Lat, el.Lon
			if el.Type != "node" {
				plLat = el.Center.Lat
				plLon = el.Center.Lon
			}
			name := el.Tags["name"]
			if name == "" {
				name = "Bilinmeyen"
			}
			results = append(results, models.Place{
				Name:      name,
				Latitude:  plLat,
				Longitude: plLon,
				Category:  interest,
			})
		}
	}

	// Redis'e kaydet
	if jsonData, err := json.Marshal(results); err == nil {
		RedisClient.Set(Ctx, cacheKey, jsonData, 6*time.Hour)
	}
	if isDebug {
		fmt.Printf("[DEBUG] Returning %d places\n", len(results))
	}
	return results, nil
}

// Maps user interest keywords to Overpass tags
func mapInterestToOverpassTag(interest string) string {
	switch interest {
	case "kafe":
		return "amenity=cafe"
	case "restoran":
		return "amenity=restaurant"
	case "park":
		return "leisure=park"
	case "manzara", "gezilecek":
		return "tourism=viewpoint"
	case "müze":
		return "tourism=museum"
	case "bar":
		return "amenity=bar"
	case "oto tamircisi":
		return "shop=car_repair"
	default:
		return "amenity=restaurant"
	}
}

// Filters only places that are within 5 km of the current user location
func GetPlacesBetweenWithCurrentLocation(startLat, startLon, endLat, endLon, currentLat, currentLon float64, interests []string) ([]models.Place, error) {
	// First, get the standard route places
	places, err := GetPlacesBetween(startLat, startLon, endLat, endLon, interests)
	if err != nil {
		return nil, err
	}

	// Calculate proximity to the current location (e.g., within 5 km)
	const maxDistance = 5000.0 // in meters

	var nearbyPlaces []models.Place
	for _, p := range places {
		d := haversineDistance(currentLat, currentLon, p.Latitude, p.Longitude)
		if d <= maxDistance {
			nearbyPlaces = append(nearbyPlaces, p)
		}
	}

	return nearbyPlaces, nil
}

// haversineDistance calculates the distance in meters between two coordinates.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth's radius in meters
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
