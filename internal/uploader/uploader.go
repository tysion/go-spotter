package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tysion/spotter/internal/handler"
	"github.com/tysion/spotter/internal/overpass"
)

func UploadBatch(url string, batch []overpass.Element) error {
	body, err := json.Marshal(toCreatePOIRequest(batch))
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("upload failed with status %s", resp.Status)
	}

	return nil
}

func toCreatePOIRequest(elements []overpass.Element) []handler.CreatePOIRequest {
	var req []handler.CreatePOIRequest
	for _, elem := range elements {
		req = append(req, handler.CreatePOIRequest{
			ID:   elem.ID,
			Lat:  elem.Lat,
			Lon:  elem.Lon,
			Tags: elem.Tags,
		})
	}
	return req
}
