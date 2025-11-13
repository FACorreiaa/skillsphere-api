-- +goose Up
-- This table should already exist from your AuthService schema.
-- We'll add a few columns to it.

ALTER TABLE users ADD COLUMN bio TEXT;
ALTER TABLE users ADD COLUMN location_text VARCHAR(255); -- For display purposes
ALTER TABLE users ADD COLUMN location_geom GEOMETRY(Point, 4326); -- For spatial queries (requires PostGIS)
ALTER TABLE users ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ; -- For soft deletes

-- Create a spatial index for fast location-based searches
CREATE INDEX idx_users_location_geom ON users USING GIST (location_geom);
