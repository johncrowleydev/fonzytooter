-- +goose Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE,
    display_name TEXT NOT NULL,
    password_hash BLOB,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (length(trim(id)) > 0),
    CHECK (username IS NULL OR length(trim(username)) > 0),
    CHECK (length(trim(display_name)) > 0),
    CHECK ((username IS NULL) = (password_hash IS NULL))
);

-- The personal deployment has one durable application owner. Credentials are
-- provisioned separately from environment configuration; this row is not an
-- anonymous or guest identity. Its stable ID is also the deterministic owner
-- for the existing single-learner data migration in the next stacked PR.
INSERT INTO users (id, username, display_name, password_hash, created_at, updated_at)
VALUES (
    '00000000-0000-4000-8000-000000000001',
    NULL,
    'Owner',
    NULL,
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
);

CREATE TABLE auth_sessions (
    token_hash TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CHECK (length(token_hash) = 64),
    CHECK (length(trim(user_id)) > 0)
);

CREATE INDEX auth_sessions_user_expires
    ON auth_sessions (user_id, expires_at);

-- +goose Down
DROP INDEX auth_sessions_user_expires;
DROP TABLE auth_sessions;
DROP TABLE users;
