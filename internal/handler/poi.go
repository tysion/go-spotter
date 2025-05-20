package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/tysion/spotter/internal/db"
	"github.com/tysion/spotter/internal/model"
	"github.com/uber/h3-go/v4"
)

type POIHandler struct {
	db         *db.DB
	resolution int
	kRing      int
}

type CreatePOIRequest struct {
	Type string         `json:"type"`
	ID   int64          `json:"id"`
	Lat  float64        `json:"lat"`
	Lon  float64        `json:"lon"`
	Tags map[string]any `json:"tags"`
}

func NewPOIHandler(db *db.DB) (*POIHandler, error) {
	return &POIHandler{db: db, resolution: 9, kRing: 1}, nil
}

func (h *POIHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *POIHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	if latStr == "" || lonStr == "" {
		log.Warn().Msg("Missing lat or lon")
		http.Error(w, "Missing lat or lon", http.StatusBadRequest)
		return
	}
	log.Info().
		Str("lat", latStr).
		Str("lon", lonStr).
		Msg("Incoming GET request")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || lat < -90 || lat > 90 {
		log.Warn().Err(err).Msg("Invalid lat")
		http.Error(w, "Invalid lat", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil || lon < -180 || lon > 180 {
		log.Warn().Err(err).Msg("Invalid lon")
		http.Error(w, "Invalid lon", http.StatusBadRequest)
		return
	}

	cell, err := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lon}, h.resolution)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert lat/lon to H3 cell")
		http.Error(w, "H3 error", http.StatusInternalServerError)
		return
	}

	cells, err := h3.GridDisk(cell, h.kRing)
	if err != nil {
		log.Error().Err(err).Msg("Failed to calculate H3 grid disk")
		http.Error(w, "h3 error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	pois, err := h.db.FindPOIsByH3Cells(ctx, cells)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query POIs by H3 cells")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(pois); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON")
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}

	log.Info().Int("result_count", len(pois)).Msg("Found POIs")
}

func (h *POIHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Incoming POST request")

	var req []CreatePOIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Invalid JSON")
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	var pois []model.POI
	for _, item := range req {
		if item.Lat < -90 || item.Lat > 90 || item.Lon < -180 || item.Lon > 180 {
			log.Error().
				Int64("ID", item.ID).
				Float64("lat", item.Lat).
				Float64("lon", item.Lon).
				Msg("Invalid coordinates")
			continue
		}

		cell, err := h3.LatLngToCell(h3.LatLng{Lat: item.Lat, Lng: item.Lon}, h.resolution)
		if err != nil {
			log.Error().
				Int64("ID", item.ID).
				Float64("lat", item.Lat).
				Float64("lon", item.Lon).
				Err(err).
				Msg("Failed to convert lat/lon to H3 cell")
				continue
		}

		amenity, ok := item.Tags["amenity"].(string)
		if !ok {
			log.Error().
				Int64("ID", item.ID).
				Float64("lat", item.Lat).
				Float64("lon", item.Lon).
				Msg("Missing amenity")
			continue
		}

		name, ok := item.Tags["name"].(string)
		if !ok {
			name = "Unknown"
		}

		pois = append(pois, model.POI{
			ID:      item.ID,
			Name:    name,
			Amenity: amenity,
			Lat:     item.Lat,
			Lon:     item.Lon,
			Cell:    cell,
			Tags:    item.Tags,
		})
	}

	ctx := r.Context()
	if err := h.db.InsertPOIsBatch(ctx, pois); err != nil {
		log.Error().Err(err).Msg("Database error")
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"ok"}`))

	log.Info().Int("rows_affected", len(pois)).Msg("Inserted POIs")
}
