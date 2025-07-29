package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"routy/models"
	"strings"
	"time"
)

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

	startCoord := []float64{startLat, startLon}
	endCoord := []float64{endLat, endLon}

	for i := 0; i < len(coordinates); i += step {
		lon := coordinates[i][0]
		lat := coordinates[i][1]

		for _, interest := range interests {
			places, err := GetPlacesNearby(lat, lon, interest, startCoord, endCoord)
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
	}

	filteredPlaces := FilterStopsByDistance(coordinates, allPlaces, 3000)
	if len(filteredPlaces) > 15 {
		filteredPlaces = filteredPlaces[:15]
	}
	return filteredPlaces, nil
}

// Fetches places near a specific coordinate, only within allowed route bounds
func GetPlacesNearby(lat, lon float64, interest string, startCoord, endCoord []float64) ([]models.Place, error) {
	var all []models.Place

	minLat := math.Min(startCoord[0], endCoord[0]) - 0.1
	maxLat := math.Max(startCoord[0], endCoord[0]) + 0.1
	minLon := math.Min(startCoord[1], endCoord[1]) - 0.1
	maxLon := math.Max(startCoord[1], endCoord[1]) + 0.1
	// Check if point is within bounding box of the route
	if lat < minLat || lat > maxLat || lon < minLon || lon > maxLon {
		log.Printf("Coordinate out of route bounds: lat=%f, lon=%f", lat, lon)
		return all, nil
	}

	osmPlaces, err := GetPlacesFromOverpass(lat, lon, interest)
	if err != nil {
		log.Printf("Overpass error: %v", err)
	} else {
		all = append(all, osmPlaces...)
	}

	return all, nil
}

// Makes Overpass API call or returns from Redis cache if available
func GetPlacesFromOverpass(lat, lon float64, interest string) ([]models.Place, error) {
	tag := mapInterestToOverpassTag(interest)
	cacheKey := fmt.Sprintf("overpass:%s:%.4f:%.4f", interest, lat, lon)

	// Check Redis cache
	cached, err := RedisClient.Get(Ctx, cacheKey).Result()
	if err == nil {
		fmt.Println("Data retrieved from Redis.")
		var results []models.Place
		if err := json.Unmarshal([]byte(cached), &results); err == nil {
			return results, nil
		}
	} else {
		fmt.Println("No data in Redis, querying Overpass API...")
	}

	// Overpass API call
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
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("Data retrieved from Overpass API.")

	body, _ := io.ReadAll(resp.Body)
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
		return nil, err
	}

	var results []models.Place
	for _, el := range data.Elements {
		lat := el.Lat
		lon := el.Lon
		if el.Type != "node" {
			lat = el.Center.Lat
			lon = el.Center.Lon
		}
		name := el.Tags["name"]
		if name == "" {
			name = "Bilinmeyen"
		}
		results = append(results, models.Place{
			Name:      name,
			Latitude:  lat,
			Longitude: lon,
			Category:  interest,
		})
	}

	jsonData, err := json.Marshal(results)
	if err == nil {
		RedisClient.Set(Ctx, cacheKey, jsonData, 6*time.Hour)
	} else {
		fmt.Println("Failed to marshal data for Redis:", err)
	}

	return results, nil
}

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
