-- Hotels Schema
-- Migration: 000002
-- Description: Create hotels and related tables

CREATE TABLE IF NOT EXISTS hotels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_code VARCHAR(100) NOT NULL,
    provider_hotel_id VARCHAR(100) NOT NULL,

    -- Hotel information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    star_rating DECIMAL(3,1),

    -- Location
    country_code VARCHAR(2) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),

    -- Images
    images JSONB,

    -- Amenities
    amenities JSONB,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Soft delete
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hotel_id UUID NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    provider_room_code VARCHAR(100) NOT NULL,

    -- Room information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    room_type VARCHAR(50),
    max_occupancy INTEGER NOT NULL,
    number_of_beds INTEGER,

    -- Amenities
    amenities JSONB,

    -- Images
    images JSONB,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    hotel_id UUID NOT NULL REFERENCES hotels(id),
    room_id UUID NOT NULL REFERENCES rooms(id),

    -- Booking details
    provider_booking_reference VARCHAR(100),
    check_in DATE NOT NULL,
    check_out DATE NOT NULL,
    number_of_guests INTEGER NOT NULL,
    number_of_rooms INTEGER DEFAULT 1,

    -- Pricing
    total_amount INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',

    -- Status (using state machine)
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- Guest information
    guest_first_name VARCHAR(100),
    guest_last_name VARCHAR(100),
    guest_email VARCHAR(255),
    guest_phone VARCHAR(20),

    -- Special requests
    special_requests TEXT,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for hotels
CREATE INDEX idx_hotels_provider_hotel_id ON hotels(provider_hotel_id);
CREATE INDEX idx_hotels_city ON hotels(city);
CREATE INDEX idx_hotels_country_code ON hotels(country_code);
CREATE INDEX idx_hotels_deleted_at ON hotels(deleted_at) WHERE deleted_at IS NOT NULL;

-- Indexes for rooms
CREATE INDEX idx_rooms_hotel_id ON rooms(hotel_id);
CREATE INDEX idx_rooms_provider_room_code ON rooms(provider_room_code);
CREATE INDEX idx_rooms_deleted_at ON rooms(deleted_at) WHERE deleted_at IS NOT NULL;

-- Indexes for bookings
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_hotel_id ON bookings(hotel_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_check_in ON bookings(check_in);
CREATE INDEX idx_bookings_created_at ON bookings(created_at);
CREATE INDEX idx_bookings_deleted_at ON bookings(deleted_at) WHERE deleted_at IS NOT NULL;

-- Triggers
CREATE TRIGGER update_hotels_updated_at BEFORE UPDATE ON hotels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rooms_updated_at BEFORE UPDATE ON rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bookings_updated_at BEFORE UPDATE ON bookings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE hotels IS 'Hotel information from various providers';
COMMENT ON TABLE rooms IS 'Room types available in hotels';
COMMENT ON TABLE bookings IS 'Booking records';
COMMENT ON COLUMN bookings.status IS 'Booking status: pending, confirmed, cancelled, etc.';
COMMENT ON COLUMN hotels.provider_code IS 'Provider identifier: hotelbeds, hotelplanner, etc.';
