CREATE TABLE IF NOT EXISTS comments (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    body text NOT NULL,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    movie_id bigint NOT NULL REFERENCES movies ON DELETE CASCADE
);