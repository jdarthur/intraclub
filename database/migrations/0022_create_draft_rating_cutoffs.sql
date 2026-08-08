-- 0022_create_draft_rating_cutoffs.sql
-- The draft_rating_cutoff join table, matching the DraftRatingCutoff record
-- shape (model/draft_rating_cutoff.go). This is the normalized replacement for
-- the former inline Draft.RatingCutoffs map (rating -> cutoff index).
-- Table name equals record.Type() ("draft_rating_cutoff").
--   id            -> RecordId hex TEXT primary key
--   draft_id      -> DraftId hex TEXT
--   rating_id     -> RatingId hex TEXT
--   cutoff_index  -> INTEGER
-- One cutoff per (draft, rating): UNIQUE(draft_id, rating_id) mirrors
-- DraftRatingCutoff.UniquenessEquivalent.
CREATE TABLE draft_rating_cutoff (
    id           TEXT PRIMARY KEY,   -- RecordId hex string
    draft_id     TEXT NOT NULL,      -- DraftId hex string
    rating_id    TEXT NOT NULL,      -- RatingId hex string
    cutoff_index INTEGER NOT NULL,
    UNIQUE (draft_id, rating_id)
);
