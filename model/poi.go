package model

type POI struct {
	ID      int64                  `json:"id"`
	Name    string                 `json:"name"`
	Amenity string                 `json:"amenity"`
	Lat     float64                `json:"lat"`
	Lon     float64                `json:"lon"`
	Cell    int64                  `json:"cell"`
	Tags    map[string]any         `json:"tags"`
}
