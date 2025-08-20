-- +goose Up
CREATE TABLE feeds (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID NOT NULL,
    url TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT fk_feeds_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;
