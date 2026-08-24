-- +goose Up
CREATE TABLE tutor_conversation_memory (
    conversation_id TEXT PRIMARY KEY REFERENCES tutor_conversations (id) ON DELETE CASCADE,
    summary TEXT NOT NULL,
    structured_json TEXT NOT NULL CHECK (json_valid(structured_json)),
    summarized_through_sequence INTEGER NOT NULL CHECK (summarized_through_sequence > 0),
    format_version INTEGER NOT NULL CHECK (format_version > 0),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE tutor_conversation_memory;
