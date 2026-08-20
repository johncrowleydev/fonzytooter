-- +goose Up
CREATE TABLE review_cards (
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
    PRIMARY KEY (course_id, module_id, review_item_id)
);

CREATE INDEX review_cards_course_due
    ON review_cards (course_id, due_at, module_id, review_item_id);

CREATE TABLE review_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    FOREIGN KEY (course_id, module_id, review_item_id)
        REFERENCES review_cards (course_id, module_id, review_item_id)
);

CREATE INDEX review_logs_card_reviewed_at_id
    ON review_logs (course_id, module_id, review_item_id, reviewed_at, id);

-- +goose Down
DROP INDEX review_logs_card_reviewed_at_id;
DROP TABLE review_logs;
DROP INDEX review_cards_course_due;
DROP TABLE review_cards;
