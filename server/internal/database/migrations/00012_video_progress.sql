-- +goose Up
CREATE TABLE video_progress (
    user_id TEXT NOT NULL REFERENCES users (id),
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    video_id TEXT NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0, 1)),
    completed_at TEXT,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (user_id, course_id, module_id, video_id),
    CHECK (completed = 1 OR completed_at IS NULL)
);

ALTER TABLE activities ADD COLUMN video_id TEXT;

-- +goose Down
ALTER TABLE activities DROP COLUMN video_id;
DROP TABLE video_progress;
