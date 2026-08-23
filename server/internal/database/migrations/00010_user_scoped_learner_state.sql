-- +goose Up
PRAGMA defer_foreign_keys = ON;

DROP TRIGGER tutor_messages_validate_continuation_insert;
DROP TRIGGER tutor_messages_validate_continuation_update;

ALTER TABLE lesson_progress RENAME TO lesson_progress_legacy;
ALTER TABLE activities RENAME TO activities_legacy;
ALTER TABLE exercise_workspaces RENAME TO exercise_workspaces_legacy;
ALTER TABLE exercise_attempts RENAME TO exercise_attempts_legacy;
ALTER TABLE exercise_test_results RENAME TO exercise_test_results_legacy;
ALTER TABLE review_cards RENAME TO review_cards_legacy;
ALTER TABLE review_logs RENAME TO review_logs_legacy;
ALTER TABLE tutor_conversations RENAME TO tutor_conversations_legacy;
ALTER TABLE tutor_messages RENAME TO tutor_messages_legacy;
ALTER TABLE tutor_message_parts RENAME TO tutor_message_parts_legacy;
ALTER TABLE tutor_tool_calls RENAME TO tutor_tool_calls_legacy;
ALTER TABLE tutor_conversation_memory RENAME TO tutor_conversation_memory_legacy;

DROP INDEX activities_course_occurred_at_id;
DROP INDEX exercise_attempts_exercise_created_at_id;
DROP INDEX review_cards_course_due;
DROP INDEX review_logs_card_reviewed_at_id;
DROP INDEX tutor_conversations_updated_at_id;
DROP INDEX tutor_messages_conversation_sequence;
DROP INDEX tutor_tool_calls_conversation_message_sequence;

CREATE TABLE lesson_progress (
    user_id TEXT NOT NULL REFERENCES users (id),
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    lesson_id TEXT NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0, 1)),
    completed_at TEXT,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (user_id, course_id, module_id, lesson_id),
    CHECK (completed = 1 OR completed_at IS NULL)
);

CREATE TABLE activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users (id),
    kind TEXT NOT NULL CHECK (length(kind) > 0),
    course_id TEXT NOT NULL,
    module_id TEXT,
    lesson_id TEXT,
    exercise_id TEXT,
    review_item_id TEXT,
    occurred_at TEXT NOT NULL
);
CREATE INDEX activities_user_course_occurred_at_id
    ON activities (user_id, course_id, occurred_at DESC, id DESC);

CREATE TABLE exercise_workspaces (
    user_id TEXT NOT NULL REFERENCES users (id),
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    exercise_id TEXT NOT NULL,
    code TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (user_id, course_id, module_id, exercise_id)
);

CREATE TABLE exercise_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users (id),
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    exercise_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    passed_count INTEGER NOT NULL CHECK (passed_count >= 0),
    failed_count INTEGER NOT NULL CHECK (failed_count >= 0),
    duration_ms INTEGER NOT NULL CHECK (duration_ms >= 0),
    all_passed INTEGER NOT NULL CHECK (all_passed IN (0, 1)),
    code_snapshot TEXT NOT NULL,
    UNIQUE (user_id, id)
);
CREATE INDEX exercise_attempts_user_exercise_created_at_id
    ON exercise_attempts (user_id, course_id, module_id, exercise_id, created_at DESC, id DESC);

CREATE TABLE exercise_test_results (
    user_id TEXT NOT NULL,
    attempt_id INTEGER NOT NULL,
    test_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('passed', 'failed', 'error')),
    message TEXT NOT NULL,
    duration_ms INTEGER NOT NULL CHECK (duration_ms >= 0),
    PRIMARY KEY (user_id, attempt_id, test_id),
    FOREIGN KEY (user_id, attempt_id) REFERENCES exercise_attempts (user_id, id) ON DELETE CASCADE
);

CREATE TABLE review_cards (
    user_id TEXT NOT NULL REFERENCES users (id),
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    review_item_id TEXT NOT NULL,
    due_at TEXT NOT NULL,
    stability REAL NOT NULL,
    difficulty REAL NOT NULL,
    scheduled_days INTEGER NOT NULL CHECK (scheduled_days >= 0),
    reps INTEGER NOT NULL CHECK (reps >= 0),
    lapses INTEGER NOT NULL CHECK (lapses >= 0),
    state INTEGER NOT NULL CHECK (state BETWEEN 0 AND 3),
    last_review_at TEXT,
    remaining_steps INTEGER NOT NULL CHECK (remaining_steps >= 0),
    updated_at TEXT NOT NULL,
    PRIMARY KEY (user_id, course_id, module_id, review_item_id)
);
CREATE INDEX review_cards_user_course_due
    ON review_cards (user_id, course_id, due_at, module_id, review_item_id);

CREATE TABLE review_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users (id),
    course_id TEXT NOT NULL,
    module_id TEXT NOT NULL,
    review_item_id TEXT NOT NULL,
    reviewed_at TEXT NOT NULL,
    rating TEXT NOT NULL CHECK (rating IN ('again', 'hard', 'good', 'easy')),
    previous_due TEXT NOT NULL,
    next_due TEXT NOT NULL,
    before_stability REAL NOT NULL,
    after_stability REAL NOT NULL,
    before_difficulty REAL NOT NULL,
    after_difficulty REAL NOT NULL,
    before_scheduled_days INTEGER NOT NULL CHECK (before_scheduled_days >= 0),
    after_scheduled_days INTEGER NOT NULL CHECK (after_scheduled_days >= 0),
    before_reps INTEGER NOT NULL CHECK (before_reps >= 0),
    after_reps INTEGER NOT NULL CHECK (after_reps >= 0),
    before_lapses INTEGER NOT NULL CHECK (before_lapses >= 0),
    after_lapses INTEGER NOT NULL CHECK (after_lapses >= 0),
    before_state INTEGER NOT NULL CHECK (before_state BETWEEN 0 AND 3),
    after_state INTEGER NOT NULL CHECK (after_state BETWEEN 0 AND 3),
    before_last_review_at TEXT,
    after_last_review_at TEXT,
    before_remaining_steps INTEGER NOT NULL CHECK (before_remaining_steps >= 0),
    after_remaining_steps INTEGER NOT NULL CHECK (after_remaining_steps >= 0),
    FOREIGN KEY (user_id, course_id, module_id, review_item_id)
        REFERENCES review_cards (user_id, course_id, module_id, review_item_id)
);
CREATE INDEX review_logs_user_card_reviewed_at_id
    ON review_logs (user_id, course_id, module_id, review_item_id, reviewed_at, id);

CREATE TABLE tutor_conversations (
    id TEXT PRIMARY KEY CHECK (length(id) > 0),
    user_id TEXT NOT NULL REFERENCES users (id),
    course_id TEXT,
    title TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (user_id, id),
    CHECK (course_id IS NULL OR length(course_id) > 0)
);
CREATE INDEX tutor_conversations_user_updated_at_id
    ON tutor_conversations (user_id, updated_at DESC, id DESC);

CREATE TABLE tutor_messages (
    id TEXT PRIMARY KEY CHECK (length(id) > 0),
    user_id TEXT NOT NULL,
    conversation_id TEXT NOT NULL,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    created_at TEXT NOT NULL,
    continuation_provider TEXT,
    continuation_model TEXT,
    continuation_state_json TEXT,
    UNIQUE (user_id, conversation_id, sequence),
    UNIQUE (user_id, conversation_id, id),
    UNIQUE (user_id, id),
    FOREIGN KEY (user_id, conversation_id) REFERENCES tutor_conversations (user_id, id) ON DELETE CASCADE
);
CREATE INDEX tutor_messages_user_conversation_sequence
    ON tutor_messages (user_id, conversation_id, sequence);

CREATE TABLE tutor_message_parts (
    user_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    kind TEXT NOT NULL CHECK (kind = 'text'),
    text_content TEXT NOT NULL,
    PRIMARY KEY (user_id, message_id, sequence),
    FOREIGN KEY (user_id, message_id) REFERENCES tutor_messages (user_id, id) ON DELETE CASCADE
);

CREATE TABLE tutor_tool_calls (
    id TEXT PRIMARY KEY CHECK (length(id) > 0),
    user_id TEXT NOT NULL,
    conversation_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    request_id TEXT NOT NULL CHECK (length(request_id) > 0),
    name TEXT NOT NULL CHECK (length(name) > 0),
    arguments_json TEXT NOT NULL CHECK (json_valid(arguments_json)),
    status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'failed')),
    result_json TEXT CHECK (result_json IS NULL OR json_valid(result_json)),
    error TEXT,
    created_at TEXT NOT NULL,
    completed_at TEXT,
    UNIQUE (user_id, message_id, sequence),
    UNIQUE (user_id, message_id, request_id),
    UNIQUE (user_id, id),
    FOREIGN KEY (user_id, conversation_id, message_id)
        REFERENCES tutor_messages (user_id, conversation_id, id) ON DELETE CASCADE,
    CHECK (
        (status = 'pending' AND result_json IS NULL AND error IS NULL AND completed_at IS NULL)
        OR (status = 'completed' AND error IS NULL AND completed_at IS NOT NULL)
        OR (status = 'failed' AND result_json IS NULL AND error IS NOT NULL AND completed_at IS NOT NULL)
    )
);
CREATE INDEX tutor_tool_calls_user_conversation_message_sequence
    ON tutor_tool_calls (user_id, conversation_id, message_id, sequence);

CREATE TABLE tutor_conversation_memory (
    user_id TEXT NOT NULL,
    conversation_id TEXT NOT NULL,
    summary TEXT NOT NULL,
    structured_json TEXT NOT NULL CHECK (json_valid(structured_json)),
    summarized_through_sequence INTEGER NOT NULL CHECK (summarized_through_sequence > 0),
    format_version INTEGER NOT NULL CHECK (format_version > 0),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (user_id, conversation_id),
    FOREIGN KEY (user_id, conversation_id) REFERENCES tutor_conversations (user_id, id) ON DELETE CASCADE
);

INSERT INTO lesson_progress SELECT '00000000-0000-4000-8000-000000000001', * FROM lesson_progress_legacy;
INSERT INTO activities (id, user_id, kind, course_id, module_id, lesson_id, exercise_id, review_item_id, occurred_at)
    SELECT id, '00000000-0000-4000-8000-000000000001', kind, course_id, module_id, lesson_id, exercise_id, review_item_id, occurred_at FROM activities_legacy;
INSERT INTO exercise_workspaces SELECT '00000000-0000-4000-8000-000000000001', * FROM exercise_workspaces_legacy;
INSERT INTO exercise_attempts (id, user_id, course_id, module_id, exercise_id, created_at, passed_count, failed_count, duration_ms, all_passed, code_snapshot)
    SELECT id, '00000000-0000-4000-8000-000000000001', course_id, module_id, exercise_id, created_at, passed_count, failed_count, duration_ms, all_passed, code_snapshot FROM exercise_attempts_legacy;
INSERT INTO exercise_test_results (user_id, attempt_id, test_id, status, message, duration_ms)
    SELECT '00000000-0000-4000-8000-000000000001', attempt_id, test_id, status, message, duration_ms FROM exercise_test_results_legacy;
INSERT INTO review_cards SELECT '00000000-0000-4000-8000-000000000001', * FROM review_cards_legacy;
INSERT INTO review_logs (id, user_id, course_id, module_id, review_item_id, reviewed_at, rating, previous_due, next_due,
    before_stability, after_stability, before_difficulty, after_difficulty, before_scheduled_days, after_scheduled_days,
    before_reps, after_reps, before_lapses, after_lapses, before_state, after_state, before_last_review_at,
    after_last_review_at, before_remaining_steps, after_remaining_steps)
    SELECT id, '00000000-0000-4000-8000-000000000001', course_id, module_id, review_item_id, reviewed_at, rating,
    previous_due, next_due, before_stability, after_stability, before_difficulty, after_difficulty,
    before_scheduled_days, after_scheduled_days, before_reps, after_reps, before_lapses, after_lapses,
    before_state, after_state, before_last_review_at, after_last_review_at, before_remaining_steps,
    after_remaining_steps FROM review_logs_legacy;
INSERT INTO tutor_conversations (id, user_id, course_id, title, created_at, updated_at)
    SELECT id, '00000000-0000-4000-8000-000000000001', course_id, title, created_at, updated_at FROM tutor_conversations_legacy;
INSERT INTO tutor_messages (id, user_id, conversation_id, sequence, role, created_at, continuation_provider, continuation_model, continuation_state_json)
    SELECT id, '00000000-0000-4000-8000-000000000001', conversation_id, sequence, role, created_at, continuation_provider, continuation_model, continuation_state_json FROM tutor_messages_legacy;
INSERT INTO tutor_message_parts (user_id, message_id, sequence, kind, text_content)
    SELECT '00000000-0000-4000-8000-000000000001', message_id, sequence, kind, text_content FROM tutor_message_parts_legacy;
INSERT INTO tutor_tool_calls (id, user_id, conversation_id, message_id, sequence, request_id, name, arguments_json, status, result_json, error, created_at, completed_at)
    SELECT id, '00000000-0000-4000-8000-000000000001', conversation_id, message_id, sequence, request_id, name, arguments_json, status, result_json, error, created_at, completed_at FROM tutor_tool_calls_legacy;
INSERT INTO tutor_conversation_memory (user_id, conversation_id, summary, structured_json, summarized_through_sequence, format_version, created_at, updated_at)
    SELECT '00000000-0000-4000-8000-000000000001', conversation_id, summary, structured_json, summarized_through_sequence, format_version, created_at, updated_at FROM tutor_conversation_memory_legacy;

DROP TABLE tutor_conversation_memory_legacy;
DROP TABLE tutor_tool_calls_legacy;
DROP TABLE tutor_message_parts_legacy;
DROP TABLE tutor_messages_legacy;
DROP TABLE tutor_conversations_legacy;
DROP TABLE review_logs_legacy;
DROP TABLE review_cards_legacy;
DROP TABLE exercise_test_results_legacy;
DROP TABLE exercise_attempts_legacy;
DROP TABLE exercise_workspaces_legacy;
DROP TABLE activities_legacy;
DROP TABLE lesson_progress_legacy;

-- +goose StatementBegin
CREATE TRIGGER tutor_messages_validate_continuation_insert
BEFORE INSERT ON tutor_messages
WHEN NOT (
    (NEW.continuation_provider IS NULL AND NEW.continuation_model IS NULL AND NEW.continuation_state_json IS NULL)
    OR (
        NEW.role IS 'assistant'
        AND COALESCE(length(trim(NEW.continuation_provider)), 0) > 0
        AND COALESCE(length(trim(NEW.continuation_model)), 0) > 0
        AND NEW.continuation_state_json IS NOT NULL
    )
)
BEGIN
    SELECT RAISE(ABORT, 'invalid tutor provider continuation state');
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER tutor_messages_validate_continuation_update
BEFORE UPDATE OF role, continuation_provider, continuation_model, continuation_state_json ON tutor_messages
WHEN NOT (
    (NEW.continuation_provider IS NULL AND NEW.continuation_model IS NULL AND NEW.continuation_state_json IS NULL)
    OR (
        NEW.role IS 'assistant'
        AND COALESCE(length(trim(NEW.continuation_provider)), 0) > 0
        AND COALESCE(length(trim(NEW.continuation_model)), 0) > 0
        AND NEW.continuation_state_json IS NOT NULL
    )
)
BEGIN
    SELECT RAISE(ABORT, 'invalid tutor provider continuation state');
END;
-- +goose StatementEnd

-- +goose Down
-- User ownership is intentionally irreversible: a down migration could not
-- safely collapse independent users without discarding or conflating data.
SELECT 1;
