-- 0017_create_drafts.sql
-- The draft table, matching the Draft record shape (model/draft.go).
-- Table name equals record.Type() ("draft").
--   id                  -> DraftId hex TEXT primary key
--   name                -> TEXT
--   owner               -> UserId hex TEXT
--   format              -> FormatId hex TEXT
--   completed_at        -> RFC3339 TEXT
--   draft_order_pattern -> TEXT (DraftOrderPattern interface, stored by Name())
CREATE TABLE draft (
    id                  TEXT PRIMARY KEY,   -- DraftId hex string
    name                TEXT NOT NULL,
    owner               TEXT NOT NULL,      -- UserId hex string
    format              TEXT NOT NULL,      -- FormatId hex string
    completed_at        TEXT NOT NULL,      -- RFC3339
    draft_order_pattern TEXT NOT NULL       -- DraftOrderPattern.Name()
);
