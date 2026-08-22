-- +goose Up
CREATE TABLE tutor_conversations (
    id TEXT PRIMARY KEY CHECK (length(id) > 0),
    course_id TEXT,
    title TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    CHECK (course_id IS NULL OR length(course_id) > 0)
);

CREATE INDEX tutor_conversations_updated_at_id
    ON tutor_conversations (updated_at DESC, id DESC);

CREATE TABLE tutor_messages (
    id TEXT PRIMARY KEY CHECK (length(id) > 0),
    conversation_id TEXT NOT NULL REFERENCES tutor_conversations (id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    created_at TEXT NOT NULL,
    UNIQUE (conversation_id, sequence),
    UNIQUE (conversation_id, id)
);

CREATE INDEX tutor_messages_conversation_sequence
    ON tutor_messages (conversation_id, sequence);

CREATE TABLE tutor_message_parts (
    message_id TEXT NOT NULL REFERENCES tutor_messages (id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    kind TEXT NOT NULL CHECK (kind = 'text'),
    text_content TEXT NOT NULL,
    PRIMARY KEY (message_id, sequence)
);

CREATE TABLE tutor_tool_calls (
    id TEXT PRIMARY KEY CHECK (length(id) > 0),
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
    UNIQUE (message_id, sequence),
    UNIQUE (message_id, request_id),
    FOREIGN KEY (conversation_id, message_id)
        REFERENCES tutor_messages (conversation_id, id) ON DELETE CASCADE,
    CHECK (
        (status = 'pending' AND result_json IS NULL AND error IS NULL AND completed_at IS NULL)
        OR (status = 'completed' AND error IS NULL AND completed_at IS NOT NULL)
        OR (status = 'failed' AND result_json IS NULL AND error IS NOT NULL AND completed_at IS NOT NULL)
    )
);

CREATE INDEX tutor_tool_calls_conversation_message_sequence
    ON tutor_tool_calls (conversation_id, message_id, sequence);

-- +goose Down
DROP INDEX tutor_tool_calls_conversation_message_sequence;
DROP TABLE tutor_tool_calls;
DROP TABLE tutor_message_parts;
DROP INDEX tutor_messages_conversation_sequence;
DROP TABLE tutor_messages;
DROP INDEX tutor_conversations_updated_at_id;
DROP TABLE tutor_conversations;
