CREATE TABLE movies (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    poster_image_url VARCHAR(500),
    genre VARCHAR(100)
);

CREATE INDEX idx_movies_deleted_at ON movies(deleted_at);
