-- +goose Up
CREATE TABLE lesson_progress (
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    lesson_id TEXT NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0, 1)),
    completed_at TEXT,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (course_id, module_id, lesson_id),
    CHECK (completed = 1 OR completed_at IS NULL)
);

CREATE TABLE activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    kind TEXT NOT NULL CHECK (length(kind) > 0),
    course_id TEXT NOT NULL,
    module_id TEXT,
    lesson_id TEXT,
    exercise_id TEXT,
    review_item_id TEXT,
    occurred_at TEXT NOT NULL
);

CREATE INDEX activities_course_occurred_at_id
    ON activities (course_id, occurred_at DESC, id DESC);

-- +goose Down
DROP INDEX activities_course_occurred_at_id;
DROP TABLE activities;
DROP TABLE lesson_progress;
