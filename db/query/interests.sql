-- name: ListInterests :many
SELECT id, slug, name, category
FROM personal.interests
ORDER BY category, name, id;

-- name: GetInterest :one
SELECT id, slug, name, category
FROM personal.interests
WHERE id = $1;

-- name: ListUserInterests :many
SELECT i.id, i.slug, i.name, i.category
FROM personal.user_interests AS ui
JOIN personal.interests AS i ON i.id = ui.interest_id
WHERE ui.user_id = $1
ORDER BY i.category, i.name, i.id;

-- name: AddUserInterest :exec
INSERT INTO personal.user_interests (user_id, interest_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveUserInterest :exec
DELETE FROM personal.user_interests
WHERE user_id = $1 AND interest_id = $2;
