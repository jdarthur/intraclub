-- 0019_create_draft_captains.sql
-- The draft_captain join table, matching the DraftCaptain record shape
-- (model/draft_captain.go).
-- Table name equals record.Type() ("draft_captain").
--   id          -> RecordId hex TEXT primary key
--   draft_id    -> DraftId hex TEXT
--   team_id     -> TeamId hex TEXT
--   captain_id  -> UserId hex TEXT
--   draft_order -> INTEGER
--   created_at  -> RFC3339 TEXT
--   updated_at  -> RFC3339 TEXT
CREATE TABLE draft_captain (
    id          TEXT PRIMARY KEY,   -- RecordId hex string
    draft_id    TEXT NOT NULL,      -- DraftId hex string
    team_id     TEXT NOT NULL,      -- TeamId hex string
    captain_id  TEXT NOT NULL,      -- UserId hex string
    draft_order INTEGER NOT NULL,
    created_at  TEXT NOT NULL,      -- RFC3339
    updated_at  TEXT NOT NULL       -- RFC3339
);
