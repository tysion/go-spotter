package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tysion/spotter/internal/batcher"
	"github.com/tysion/spotter/internal/logger"
	"github.com/tysion/spotter/internal/overpass"
	"github.com/tysion/spotter/internal/uploader"
)

func main() {
	logger.Setup()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	bbox := [4]float64{55.56003576747257, 37.25573730468751, 55.9129457648424, 37.95227050781251} // Moscow

	pois, err := overpass.FetchPOIs(bbox[:])
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to fetch POIs")
	}

	batches := batcher.Split(pois, 1024)

	endpoint := os.Getenv("SPOTTER_API_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8080/poi"
	}

	for i, b := range batches {
		if err := uploader.UploadBatch(endpoint, b); err != nil {
			log.Error().Err(err).Msg("Failed to upload a batch")
		} else {
			log.Info().Int("#i", i).Int("rows_count", len(b)).Msg("Uploaded a batch")
		}
	}

	log.Info().Msg("Upload complete")
}
