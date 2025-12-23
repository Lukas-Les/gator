-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
WHERE feeds.id = $1
ORDER BY last_fetched_at NULLS FIRST;
