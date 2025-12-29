-- Migration: AI-Powered Search Schema
-- Version: 007
-- Description: Add tables for AI-powered semantic search, natural language queries, and smart recommendations

-- Enable pgvector extension for vector operations
CREATE EXTENSION IF NOT EXISTS vector;

-- Create hotel_embeddings table for storing vector embeddings
CREATE TABLE IF NOT EXISTS hotel_embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hotel_id UUID NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    embedding vector(1536), -- OpenAI text-embedding-3-small dimension
    embedding_model VARCHAR(100) DEFAULT 'text-embedding-3-small',
    embedding_version VARCHAR(50) DEFAULT 'v1',
    language VARCHAR(10) DEFAULT 'en',
    embedding_type VARCHAR(50) DEFAULT 'description', -- description, amenities, location, combined
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_hotel_embedding UNIQUE (hotel_id, embedding_type, language)
);

CREATE INDEX idx_hotel_embeddings_embedding_idx ON hotel_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_hotel_embeddings_hotel_id ON hotel_embeddings(hotel_id);

-- Create search_history table for tracking user searches
CREATE TABLE IF NOT EXISTS search_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    query TEXT NOT NULL,
    query_type VARCHAR(50) DEFAULT 'semantic', -- semantic, traditional, natural_language
    query_params JSONB, -- Store parsed query parameters
    results_count INTEGER,
    filters_applied JSONB,
    result_ids UUID[], -- IDs of returned hotels
    clicked_hotel_id UUID REFERENCES hotels(id) ON DELETE SET NULL,
    booked_hotel_id UUID REFERENCES hotels(id) ON DELETE SET NULL,
    session_id UUID,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_search_history_user_id ON search_history(user_id);
CREATE INDEX idx_search_history_created_at ON search_history(created_at DESC);
CREATE INDEX idx_search_history_session_id ON search_history(session_id);

-- Create user_preferences table for personalization
CREATE TABLE IF NOT EXISTS user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Price preference
    preferred_price_min INTEGER,
    preferred_price_max INTEGER,

    -- Location preferences
    preferred_cities TEXT[],
    preferred_countries TEXT[],

    -- Amenity preferences
    preferred_amenities TEXT[],

    -- Hotel type preferences
    preferred_hotel_types TEXT[], -- resort, city_hotel, villa, apartment, etc

    -- Travel style
    travel_styles TEXT[], -- business, leisure, adventure, family, romantic

    -- Interaction history
    view_count INTEGER DEFAULT 0,
    search_count INTEGER DEFAULT 0,
    booking_count INTEGER DEFAULT 0,

    -- Derived preferences (from ML)
    derived_preferences JSONB,
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Create search_analytics table for tracking search performance
CREATE TABLE IF NOT EXISTS search_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date DATE NOT NULL,
    metric_type VARCHAR(50) NOT NULL, -- query_count, zero_results, click_rate, booking_rate
    metrics_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_date_metric UNIQUE (date, metric_type)
);

CREATE INDEX idx_search_analytics_date ON search_analytics(date);

-- Create recommendations table for caching recommendations
CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    recommendation_type VARCHAR(50) NOT NULL, -- personalized, trending, similar, collaborative
    source_hotel_id UUID REFERENCES hotels(id) ON DELETE CASCADE, -- For "similar hotels"
    recommended_hotel_ids UUID[] NOT NULL,
    recommendation_reason TEXT,
    score FLOAT,
    metadata JSONB, -- Additional context
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recommendations_user_id ON recommendations(user_id);
CREATE INDEX idx_recommendations_type ON recommendations(recommendation_type);
CREATE INDEX idx_recommendations_expires_at ON recommendations(expires_at);

-- Create nlp_entities table for storing named entities
CREATE TABLE IF NOT EXISTS nlp_entities (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL, -- city, country, hotel_type, amenity, landmark
    entity_name VARCHAR(255) NOT NULL,
    aliases TEXT[], -- Alternative names/spellings
    synonyms TEXT[],
    embedding vector(1536),
    metadata JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_nlp_entities_type_name ON nlp_entities(entity_type, entity_name) WHERE is_active = true;
CREATE INDEX idx_nlp_entities_embedding_idx ON nlp_entities USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_ai_search_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE TRIGGER trigger_update_hotel_embeddings_updated_at
BEFORE UPDATE ON hotel_embeddings
FOR EACH ROW
EXECUTE FUNCTION update_ai_search_updated_at();

CREATE TRIGGER trigger_update_user_preferences_updated_at
BEFORE UPDATE ON user_preferences
FOR EACH ROW
EXECUTE FUNCTION update_ai_search_updated_at();

CREATE TRIGGER trigger_update_nlp_entities_updated_at
BEFORE UPDATE ON nlp_entities
FOR EACH ROW
EXECUTE FUNCTION update_ai_search_updated_at();

-- Function to cleanup expired recommendations
CREATE OR REPLACE FUNCTION cleanup_expired_recommendations()
RETURNS void AS $$
BEGIN
    DELETE FROM recommendations
    WHERE expires_at < CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql;

-- Insert sample NLP entities for common Indonesian locations
INSERT INTO nlp_entities (entity_type, entity_name, aliases, synonyms, metadata) VALUES
('city', 'Jakarta', ARRAY['DKI Jakarta', 'Batavia'], ARRAY['indonesia capital city'], '{"country": "Indonesia", "country_code": "ID"}'),
('city', 'Bali', ARRAY['Denpasar', 'Badung'], ARRAY['island', 'paradise'], '{"country": "Indonesia", "country_code": "ID", "province": "Bali"}'),
('city', 'Surabaya', ARRAY['Hero City'], ARRAY['second city'], '{"country": "Indonesia", "country_code": "ID"}'),
('city', 'Bandung', ARRAY['Kota Kembang'], ARRAY['flower city'], '{"country": "Indonesia", "country_code": "ID"}'),
('city', 'Medan', ARRAY[], ARRAY['third city'], '{"country": "Indonesia", "country_code": "ID"}'),
('city', 'Yogyakarta', ARRAY['Jogja', 'Jogjakarta'], ARRAY['cultural city', 'education city'], '{"country": "Indonesia", "country_code": "ID"}'),
('hotel_type', 'resort', ARRAY['resort hotel', 'beach resort'], ARRAY['holiday', 'vacation'], '{"category": "property_type"}'),
('hotel_type', 'city_hotel', ARRAY['business hotel', 'downtown hotel'], ARRAY['urban', 'commercial'], '{"category": "property_type"}'),
('hotel_type', 'villa', ARRAY['private villa', 'luxury villa'], ARRAY['holiday home', 'vacation rental'], '{"category": "property_type"}'),
('amenity', 'swimming_pool', ARRAY['pool', 'kolam renang'], ARRAY['water facility', 'recreation'], '{"category": "facility"}'),
('amenity', 'spa', ARRAY['wellness', 'massage'], ARRAY['relaxation', 'treatment'], '{"category": "wellness"}'),
('amenity', 'wifi', ARRAY['wireless_internet', 'internet'], ARRAY['connectivity', 'network'], '{"category": "technology"}'),
('amenity', 'gym', ARRAY['fitness_center', 'fitness'], ARRAY['workout', 'exercise'], '{"category": "fitness"}'),
('amenity', 'restaurant', ARRAY['dining', 'eatery'], ARRAY['food', 'cuisine'], '{"category": "dining"}'),
('amenity', 'parking', ARRAY['car_park', 'vehicle_parking'], ARRAY['vehicle', 'transport'], '{"category": "convenience"}')
ON CONFLICT (entity_type, entity_name) DO NOTHING;

-- Create view for search analytics summary
CREATE OR REPLACE VIEW search_analytics_summary AS
SELECT
    DATE(created_at) as date,
    query_type,
    COUNT(*) as total_searches,
    AVG(results_count) as avg_results,
    COUNT(clicked_hotel_id) as total_clicks,
    COUNT(booked_hotel_id) as total_bookings,
    ROUND(COUNT(booked_hotel_id)::numeric / NULLIF(COUNT(*), 0) * 100, 2) as booking_rate
FROM search_history
WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(created_at), query_type
ORDER BY date DESC, query_type;
