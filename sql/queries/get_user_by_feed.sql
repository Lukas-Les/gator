-- name: GetUserByFeed :one
SELECT name FROM users
JOIN feeds ON feeds.user_id = users.id
WHERE feeds.id = $1
