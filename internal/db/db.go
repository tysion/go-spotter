package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tysion/spotter/internal/model"
	"github.com/uber/h3-go/v4"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(dsn string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) InsertPOIsBatch(ctx context.Context, pois []model.POI) error {
	batch := &pgx.Batch{}
	for _, poi := range pois {
		tagsJSON, err := json.Marshal(poi.Tags)
		if err != nil {
			return fmt.Errorf("failed to marshall tags %w", err)
		}
		batch.Queue(
			`INSERT INTO spotter.pois (id, name, amenity, lat, lon, cell, tags)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			poi.ID, poi.Name, poi.Amenity, poi.Lat, poi.Lon, poi.Cell, tagsJSON,
		)
	}
	br := db.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := range pois {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("failed to insert poi #%d: %w", i, err)
		}
	}

	return nil
}

func (db *DB) FindPOIsByH3Cells(ctx context.Context, cells []h3.Cell) ([]model.POI, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, amenity, lat, lon, cell, tags
		FROM spotter.pois
		WHERE cell = ANY($1)`, cells)
	if err != nil {
		return nil, fmt.Errorf("query pois: %w", err)
	}
	defer rows.Close()

	pois := make([]model.POI, 0)
	for rows.Next() {
		var p model.POI
		if err := rows.Scan(&p.ID, &p.Name, &p.Amenity, &p.Lat, &p.Lon, &p.Cell, &p.Tags); err != nil {
			return nil, fmt.Errorf("scan poi: %w", err)
		}
		pois = append(pois, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration: %w", err)
	}

	return pois, nil
}
