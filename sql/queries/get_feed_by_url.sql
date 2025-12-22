-- name: GetFeedsByUrl :one
SELECT * FROM feeds
WHERE url = $1
ORDER BY created_at DESC;
