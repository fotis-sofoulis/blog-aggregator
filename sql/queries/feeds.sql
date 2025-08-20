-- name: AddFeed :one
INSERT INTO feeds (id, name, user_id, url, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;
