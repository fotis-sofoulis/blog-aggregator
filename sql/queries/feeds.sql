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

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds SET updated_at = NOW(), last_fetched_at = NOW() WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds ORDER BY last_fetched_at ASC NULLS FIRST LIMIT 1;
