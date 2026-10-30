-- name: GetProfile :one
SELECT user_id, username, bio, avatar_url, created_at, updated_at
FROM personal.profiles
WHERE user_id = $1;

-- name: UpsertProfile :one
INSERT INTO personal.profiles (user_id, username, bio, avatar_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE SET
    username = EXCLUDED.username,
    bio = EXCLUDED.bio,
    avatar_url = EXCLUDED.avatar_url,
    updated_at = NOW()
RETURNING user_id, username, bio, avatar_url, created_at, updated_at;
