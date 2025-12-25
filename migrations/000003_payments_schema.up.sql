-- Payments Schema
-- Migration: 000003
-- Description: Create payments table

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES bookings(id),

    -- Provider information
    provider VARCHAR(50) NOT NULL,
    provider_reference VARCHAR(255),
    payment_method VARCHAR(50),

    -- Amount
    amount INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',

    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- Payment details
    payment_url TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Metadata
    metadata JSONB,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_provider_reference ON payments(provider_reference);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_deleted_at ON payments(deleted_at) WHERE deleted_at IS NOT NULL;

-- Trigger
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE payments IS 'Payment records for bookings';
COMMENT ON COLUMN payments.provider IS 'Payment provider: midtrans, stripe, xendit';
COMMENT ON COLUMN payments.status IS 'Payment status: pending, success, failed, refunded';
