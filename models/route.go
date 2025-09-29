package models

type RouteRequest struct {
	Start     string   `json:"start"`
	End       string   `json:"end"`
	Interests []string `json:"interests"` 
}

type RouteResponse struct {
	Start      string    `json:"start"`
	End        string    `json:"end"`
	Stops      []Place   `json:"stops"`
	StartCoord []float64 `json:"start_coord"`
	EndCoord   []float64 `json:"end_coord"`
}
type RouteRequestWithCurrent struct {
	Start   string `json:"start"`
	End     string `json:"end"`
	Current struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"current"`
	Interests []string `json:"interests"`
}
