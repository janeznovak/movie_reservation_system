CREATE TABLE reservations (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL REFERENCES users(id),
    showtime_id BIGINT NOT NULL REFERENCES showtimes(id),
    status VARCHAR(50) NOT NULL DEFAULT 'confirmed',
    total_price NUMERIC(10, 2) NOT NULL
);

CREATE INDEX idx_reservations_user_id ON reservations(user_id);
CREATE INDEX idx_reservations_showtime_id ON reservations(showtime_id);
CREATE INDEX idx_reservations_deleted_at ON reservations(deleted_at);

CREATE TABLE reservation_seats (
    reservation_id BIGINT NOT NULL REFERENCES reservations(id),
    seat_id        BIGINT NOT NULL REFERENCES seats(id),
    PRIMARY KEY (reservation_id, seat_id)
);
