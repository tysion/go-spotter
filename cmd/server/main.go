package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tysion/spotter/internal/db"
	"github.com/tysion/spotter/internal/handler"
	"github.com/tysion/spotter/internal/logger"
)

func main() {
	logger.Setup()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal().Msg("DB_PASSWORD not set")
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "35432"
	}

	dsn := fmt.Sprintf("postgres://spotter:%s@%s:%s/spotter", dbPassword, dbHost, dbPort)

	database, err := db.New(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize DB")
	}
	defer database.Close()

	poiHandler, err := handler.NewPOIHandler(database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize POI handler")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/poi", poiHandler.Handle)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Info().Msg("Server running on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
