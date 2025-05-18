package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tysion/spotter/db"
	"github.com/tysion/spotter/handler"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/postgres" // или своё
	}

	database, err := db.New(dsn)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	defer database.Close()

	poiHandler, err := handler.NewPOIHandler(database)
	if err != nil {
		log.Fatalf("failed to initialize POI handler: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/poi", poiHandler.GetPOI)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Server running on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}