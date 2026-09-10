-- Enum for rider availability status
CREATE TYPE rider_status AS ENUM ('available', 'busy', 'offline');

CREATE TABLE riders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    status rider_status NOT NULL DEFAULT 'offline',
    current_lat DOUBLE PRECISION,
    current_lng DOUBLE PRECISION,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);