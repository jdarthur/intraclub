-- 0036_create_match_editors.sql
-- The match_editor child table, matching the MatchEditor record shape
-- (model/individual_match.go). This is the normalized replacement for the
-- former inline IndividualMatch.Editors slice.
-- Table name equals record.Type() ("match_editor").
--   id              -> RecordId hex TEXT primary key
--   match_id        -> IndividualMatchId hex TEXT
--   editor_user_id  -> UserId hex TEXT
CREATE TABLE match_editor (
    id              TEXT PRIMARY KEY,   -- RecordId hex string
    match_id        TEXT NOT NULL,      -- IndividualMatchId hex string
    editor_user_id  TEXT NOT NULL       -- UserId hex string
);
