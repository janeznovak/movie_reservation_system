CREATE TABLE venues (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(255) NOT NULL
);

CREATE INDEX idx_venues_deleted_at ON venues(deleted_at);

CREATE TABLE seats (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    venue_id BIGINT NOT NULL REFERENCES venues(id),
    row_label VARCHAR(10) NOT NULL,
    seat_number INT NOT NULL
);

CREATE INDEX idx_seats_venue_id ON seats(venue_id);
CREATE INDEX idx_seats_deleted_at ON seats(deleted_at);
-- Prevent duplicate seats in the same venue (e.g. two "A1" seats)
CREATE UNIQUE INDEX idx_seats_unique ON seats(venue_id, row_label, seat_number) WHERE deleted_at IS NULL;
