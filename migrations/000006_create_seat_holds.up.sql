CREATE TABLE seat_holds (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    seat_id     BIGINT NOT NULL REFERENCES seats(id),
    showtime_id BIGINT NOT NULL REFERENCES showtimes(id),
    user_id     BIGINT NOT NULL REFERENCES users(id),
    expires_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_seat_holds_seat_id ON seat_holds(seat_id);
CREATE INDEX idx_seat_holds_showtime_id ON seat_holds(showtime_id);
CREATE INDEX idx_seat_holds_user_id ON seat_holds(user_id);
-- Indexed so the availability check (WHERE expires_at > NOW()) is fast
CREATE INDEX idx_seat_holds_expires_at ON seat_holds(expires_at);
CREATE INDEX idx_seat_holds_deleted_at ON seat_holds(deleted_at);
