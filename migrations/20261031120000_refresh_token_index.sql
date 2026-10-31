-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS credentials_refresh_token_hash_key
ON personal.credentials (refresh_token_hash)
WHERE refresh_token_hash IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS personal.credentials_refresh_token_hash_key;
