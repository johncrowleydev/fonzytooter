-- +goose Up
CREATE TABLE tutor_usage_windows (
    user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    window_start TEXT NOT NULL,
    reserved_turns INTEGER NOT NULL CHECK (reserved_turns >= 0),
    updated_at TEXT NOT NULL,
    PRIMARY KEY (user_id, window_start)
);

-- +goose Down
DROP TABLE tutor_usage_windows;
