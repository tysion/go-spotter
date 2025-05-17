CREATE SCHEMA IF NOT EXISTS spotter;

CREATE TYPE spotter.poi AS
(
    id BIGINT,
    name TEXT NOT NULL,
    amenity TEXT NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lon DOUBLE PRECISION NOT NULL,
    cell BIGINT NOT NULL,
    tags JSONB NOT NULL
)

CREATE TABLE IF NOT EXISTS spotter.pois (
    id BIGINT,
    name TEXT NOT NULL,
    amenity TEXT NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lon DOUBLE PRECISION NOT NULL,
    cell BIGINT NOT NULL,
    tags JSONB NOT NULL
);

CREATE INDEX idx_pois_cell on pois(cell);
CREATE INDEX idx_pois_amenity ON pois (amenity);