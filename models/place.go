package models

type Place struct {
	Name         string  `json:"name"`
	Address      string  `json:"address,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	Website      string  `json:"website,omitempty"`
	OpeningHours string  `json:"opening_hours,omitempty"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lon"`
	Category     string  `json:"category,omitempty"`
	Rating       float64 `json:"rating,omitempty"`
}
