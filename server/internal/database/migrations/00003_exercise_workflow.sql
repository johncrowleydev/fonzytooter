-- +goose Up
CREATE TABLE exercise_workspaces (
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    exercise_id TEXT NOT NULL,
    code TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (course_id, module_id, exercise_id)
);

CREATE TABLE exercise_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    exercise_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    passed_count INTEGER NOT NULL CHECK (passed_count >= 0),
    failed_count INTEGER NOT NULL CHECK (failed_count >= 0),
    duration_ms INTEGER NOT NULL CHECK (duration_ms >= 0),
    all_passed INTEGER NOT NULL CHECK (all_passed IN (0, 1)),
    code_snapshot TEXT NOT NULL
);

CREATE TABLE exercise_test_results (
    attempt_id INTEGER NOT NULL REFERENCES exercise_attempts (id) ON DELETE CASCADE,
    test_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('passed', 'failed', 'error')),
    message TEXT NOT NULL,
    duration_ms INTEGER NOT NULL CHECK (duration_ms >= 0),
    PRIMARY KEY (attempt_id, test_id)
);

CREATE INDEX exercise_attempts_exercise_created_at_id
    ON exercise_attempts (course_id, module_id, exercise_id, created_at DESC, id DESC);

-- +goose Down
DROP INDEX exercise_attempts_exercise_created_at_id;
DROP TABLE exercise_test_results;
DROP TABLE exercise_attempts;
DROP TABLE exercise_workspaces;
