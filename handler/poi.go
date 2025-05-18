package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tysion/spotter/db"
	"github.com/uber/h3-go/v4"
)

type POIHandler struct {
	db *db.DB
}

func NewPOIHandler(db *db.DB) (*POIHandler, error) {
	return &POIHandler{db: db}, nil
}

func (h *POIHandler) GetPOI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	if latStr == "" || lonStr == "" {
		http.Error(w, "Missing lat or lon", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid lat", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid lon", http.StatusBadRequest)
		return
	}

	cell, err := h3.LatLngToCell(h3.LatLng{Lat: lat, Lng: lon}, 9)
	if err != nil {
		http.Error(w, "H3 error", http.StatusInternalServerError)
		return
	}

	cells, err := h3.GridDisk(cell, 1)
	if err != nil {
		http.Error(w, "h3 error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	pois, err := h.db.FindPOIsByH3Cells(ctx, cells)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(pois); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}
}