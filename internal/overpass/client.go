package overpass

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Element struct {
	ID   int64          `json:"id"`
	Lat  float64        `json:"lat"`
	Lon  float64        `json:"lon"`
	Tags map[string]any `json:"tags"`
}

type OverpassResponse struct {
	Elements []Element `json:"elements"`
}

const OVERPASS_API_URL = "https://overpass-api.de/api/interpreter"
const OVERPASS_CONTENT_TYPE = "application/x-www-form-urlencoded"

func FetchPOIs(bbox []float64) ([]Element, error) {
	query := fmt.Sprintf(`
		[out:json][timeout:600];
		(
			node["amenity"="cafe"](%f,%f,%f,%f);
		);
		out center;`, bbox[0], bbox[1], bbox[2], bbox[3])

	resp, err := http.Post(
		OVERPASS_API_URL,
		OVERPASS_CONTENT_TYPE,
		bytes.NewBufferString(query),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result OverpassResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Elements, nil
}
