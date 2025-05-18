package model

import "github.com/uber/h3-go/v4"

type POI struct {
	ID      int64                  `json:"id"`
	Name    string                 `json:"name"`
	Amenity string                 `json:"amenity"`
	Lat     float64                `json:"lat"`
	Lon     float64                `json:"lon"`
	Cell    h3.Cell                `json:"cell"`
	Tags    map[string]any         `json:"tags"`
}
