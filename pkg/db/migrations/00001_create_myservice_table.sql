-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS myservice_records (
    id BIGSERIAL PRIMARY KEY,
    input TEXT NOT NULL,
    output TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on created_at for faster ordering
CREATE INDEX IF NOT EXISTS idx_myservice_records_created_at ON myservice_records(created_at DESC);

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_myservice_records_updated_at
    BEFORE UPDATE ON myservice_records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_myservice_records_updated_at ON myservice_records;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_myservice_records_created_at;
DROP TABLE IF EXISTS myservice_records;
-- +goose StatementEnd
