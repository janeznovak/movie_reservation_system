CREATE TABLE showtimes (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    movie_id BIGINT NOT NULL REFERENCES movies(id),
    venue_id BIGINT NOT NULL REFERENCES venues(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    price NUMERIC(10, 2) NOT NULL
);

CREATE INDEX idx_showtimes_movie_id ON showtimes(movie_id);
CREATE INDEX idx_showtimes_venue_id ON showtimes(venue_id);
-- Indexed so browsing by date is fast
CREATE INDEX idx_showtimes_start_time ON showtimes(start_time);
CREATE INDEX idx_showtimes_deleted_at ON showtimes(deleted_at);
