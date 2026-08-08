-- 0020_create_draft_formats.sql
-- The draft_format join table, matching the DraftFormat record shape
-- (model/draft_format.go).
-- Table name equals record.Type() ("draft_format").
--   id        -> RecordId hex TEXT primary key
--   draft_id  -> DraftId hex TEXT
--   format_id -> FormatId hex TEXT
-- One format per draft: UNIQUE(draft_id) mirrors DraftFormat.UniquenessEquivalent.
CREATE TABLE draft_format (
    id        TEXT PRIMARY KEY,   -- RecordId hex string
    draft_id  TEXT NOT NULL,      -- DraftId hex string
    format_id TEXT NOT NULL,      -- FormatId hex string
    UNIQUE (draft_id)
);
