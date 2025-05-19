package main

import (
	"log"
	"os"

	"github.com/tysion/spotter/internal/batcher"
	"github.com/tysion/spotter/internal/overpass"
	"github.com/tysion/spotter/internal/uploader"
)

func main() {
	bbox := [4]float64{55.56003576747257, 37.25573730468751, 55.9129457648424, 37.95227050781251} // Moscow

	pois, err := overpass.FetchPOIs(bbox[:])
	if err != nil {
		log.Fatalf("failed to fetch POIs: %v", err)
	}

	batches := batcher.Split(pois, 1024)

	endpoint := os.Getenv("SPOTTER_API_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8080/poi"
	}

	for i, b := range batches {
		if err := uploader.UploadBatch(endpoint, b); err != nil {
			log.Printf("failed to upload batch: %v", err)
		} else {
			log.Printf("uploaded batch %d\n", i)
		}
	}

	log.Println("upload complete")
}
