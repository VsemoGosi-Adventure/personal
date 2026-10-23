-- +goose Up
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE SCHEMA IF NOT EXISTS personal;

-- 1. Учетные данные и аутентификация
CREATE TABLE IF NOT EXISTS personal.credentials (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    refresh_token_hash VARCHAR(255) NULL,
    refresh_expires_at TIMESTAMPTZ NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Профили пользователей
CREATE TABLE IF NOT EXISTS personal.profiles (
    user_id UUID PRIMARY KEY REFERENCES personal.credentials(user_id) ON DELETE CASCADE,
    username VARCHAR(50) NOT NULL UNIQUE,
    bio TEXT NULL,
    avatar_url TEXT NULL,
    last_location GEOGRAPHY(Point, 4326) NULL,
    location_updated_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_profiles_last_location ON personal.profiles USING GIST (last_location);

-- 3. Справочник интересов / тегов
CREATE TABLE IF NOT EXISTS personal.interests (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Связь пользователей с интересами (Многие ко многим)
CREATE TABLE IF NOT EXISTS personal.user_interests (
    user_id UUID NOT NULL REFERENCES personal.credentials(user_id) ON DELETE CASCADE,
    interest_id INT NOT NULL REFERENCES personal.interests(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, interest_id)
);

-- +goose Down
DROP TABLE IF EXISTS personal.user_interests;
DROP TABLE IF EXISTS personal.interests;
DROP TABLE IF EXISTS personal.profiles;
DROP TABLE IF EXISTS personal.credentials;
DROP SCHEMA IF EXISTS personal CASCADE;
