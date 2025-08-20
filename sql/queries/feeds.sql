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

-- name: GetFeeds :many
SELECT f.name AS feed_name, f.url, u.name as user_name FROM feeds f JOIN users u ON f.user_id = u.id;
