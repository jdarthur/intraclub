-- 0007_create_season_late_additions.sql
-- The season_late_addition join table, matching the SeasonLateAddition record
-- shape (model/season_late_addition.go).
-- Table name equals record.Type() ("season_late_addition").
--   id         -> RecordId hex TEXT primary key
--   season_id  -> SeasonId hex TEXT
--   user_id    -> UserId hex TEXT
--   created_at -> RFC3339 TEXT
--   updated_at -> RFC3339 TEXT
CREATE TABLE season_late_addition (
    id         TEXT PRIMARY KEY,   -- RecordId hex string
    season_id  TEXT NOT NULL,      -- SeasonId hex string
    user_id    TEXT NOT NULL,      -- UserId hex string
    created_at TEXT NOT NULL,      -- RFC3339
    updated_at TEXT NOT NULL       -- RFC3339
);
