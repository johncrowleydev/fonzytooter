-- +goose Up
ALTER TABLE tutor_messages ADD COLUMN continuation_provider TEXT;
ALTER TABLE tutor_messages ADD COLUMN continuation_model TEXT;
ALTER TABLE tutor_messages ADD COLUMN continuation_state_json TEXT
    CHECK (continuation_state_json IS NULL OR json_valid(continuation_state_json));

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
DROP TRIGGER tutor_messages_validate_continuation_update;
DROP TRIGGER tutor_messages_validate_continuation_insert;
ALTER TABLE tutor_messages DROP COLUMN continuation_state_json;
ALTER TABLE tutor_messages DROP COLUMN continuation_model;
ALTER TABLE tutor_messages DROP COLUMN continuation_provider;
