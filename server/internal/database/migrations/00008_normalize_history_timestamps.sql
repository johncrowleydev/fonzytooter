-- +goose Up
-- Application-written history timestamps are UTC. Normalize legacy RFC3339Nano
-- values to the fixed-width representation used by current writes so indexed
-- TEXT ordering is chronological, including fractional seconds.
UPDATE activities
SET occurred_at = substr(occurred_at, 1, 19) || '.' ||
    CASE
        WHEN substr(occurred_at, 20, 1) = '.'
            THEN substr(substr(occurred_at, 21, length(occurred_at) - 21) || '000000000', 1, 9)
        ELSE '000000000'
    END || 'Z'
WHERE substr(occurred_at, -1) = 'Z'
  AND (length(occurred_at) = 20 OR
       (substr(occurred_at, 20, 1) = '.' AND length(occurred_at) BETWEEN 22 AND 30));

UPDATE exercise_attempts
SET created_at = substr(created_at, 1, 19) || '.' ||
    CASE
        WHEN substr(created_at, 20, 1) = '.'
            THEN substr(substr(created_at, 21, length(created_at) - 21) || '000000000', 1, 9)
        ELSE '000000000'
    END || 'Z'
WHERE substr(created_at, -1) = 'Z'
  AND (length(created_at) = 20 OR
       (substr(created_at, 20, 1) = '.' AND length(created_at) BETWEEN 22 AND 30));

UPDATE review_logs
SET reviewed_at = substr(reviewed_at, 1, 19) || '.' ||
    CASE
        WHEN substr(reviewed_at, 20, 1) = '.'
            THEN substr(substr(reviewed_at, 21, length(reviewed_at) - 21) || '000000000', 1, 9)
        ELSE '000000000'
    END || 'Z'
WHERE substr(reviewed_at, -1) = 'Z'
  AND (length(reviewed_at) = 20 OR
       (substr(reviewed_at, 20, 1) = '.' AND length(reviewed_at) BETWEEN 22 AND 30));

-- +goose Down
SELECT 1;
