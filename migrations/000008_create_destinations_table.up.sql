-- Create destinations table for syncing HotelBeds destinations
CREATE TABLE IF NOT EXISTS destinations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    country_code VARCHAR(10) NOT NULL,
    country_name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,                  -- 'CITY', 'REGION', 'COUNTRY'
    parent_code VARCHAR(50),                    -- Parent destination code
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    hotelbeds_data JSONB,                       -- Store raw API response
    synced_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for fast queries
CREATE INDEX idx_destinations_country_code ON destinations(country_code);
CREATE INDEX idx_destinations_type ON destinations(type);
CREATE INDEX idx_destinations_name_gin ON destinations USING GIN (to_tsvector('english', name));

-- Comments
COMMENT ON TABLE destinations IS 'Synced destinations from HotelBeds API';
COMMENT ON COLUMN destinations.code IS 'HotelBeds destination code (unique identifier)';
COMMENT ON COLUMN destinations.type IS 'Destination type: CITY, REGION, or COUNTRY';
COMMENT ON COLUMN destinations.hotelbeds_data IS 'Raw JSON response from HotelBeds API for debugging';

-- Trigger for updated_at (reuse existing function if exists, otherwise create it)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_destinations_updated_at
    BEFORE UPDATE ON destinations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
