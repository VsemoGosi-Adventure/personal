-- name: GetCredentialsByEmail :one
SELECT user_id, email, password_hash, role, refresh_token_hash, refresh_expires_at, is_blocked
FROM personal.credentials
WHERE email = $1;

-- name: CreateCredentials :one
INSERT INTO personal.credentials (email, password_hash, refresh_token_hash, refresh_expires_at)
VALUES ($1, $2, $3, $4)
RETURNING user_id, role;

-- name: CreateInitialProfile :exec
INSERT INTO personal.profiles (user_id, username)
VALUES ($1, $2);

-- name: UpdateRefreshToken :execrows
UPDATE personal.credentials
SET refresh_token_hash = $2, refresh_expires_at = $3, updated_at = NOW()
WHERE user_id = $1 AND NOT is_blocked;

-- name: GetCredentialsByRefreshHash :one
SELECT user_id, role, refresh_expires_at, is_blocked
FROM personal.credentials
WHERE refresh_token_hash = $1;

-- name: RotateRefreshToken :one
UPDATE personal.credentials
SET refresh_token_hash = $2, refresh_expires_at = $3, updated_at = NOW()
WHERE refresh_token_hash = $1 AND refresh_expires_at > NOW() AND NOT is_blocked
RETURNING user_id, role;

-- name: InvalidateRefreshToken :execrows
UPDATE personal.credentials
SET refresh_token_hash = NULL, refresh_expires_at = NULL, updated_at = NOW()
WHERE refresh_token_hash = $1;
