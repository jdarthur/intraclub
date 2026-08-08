-- 0016_create_weeks.sql
-- The week table, matching the Week record shape (model/week.go).
-- Table name equals record.Type() ("week").
--   id       -> WeekId hex TEXT primary key
--   draft_id -> DraftId hex TEXT
--   date     -> RFC3339 TEXT
--   note     -> TEXT
CREATE TABLE week (
    id       TEXT PRIMARY KEY,   -- WeekId hex string
    draft_id TEXT NOT NULL,      -- DraftId hex string
    date     TEXT NOT NULL,      -- RFC3339
    note     TEXT NOT NULL
);
