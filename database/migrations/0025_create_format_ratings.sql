-- 0025_create_format_ratings.sql
-- The format_rating join table, matching the FormatRating record shape
-- (model/format_rating.go). This is the normalized replacement for the former
-- inline Format.PossibleRatings slice, preserving order via RatingIndex.
-- Table name equals record.Type() ("format_rating").
--   id           -> RecordId hex TEXT primary key
--   format_id    -> FormatId hex TEXT
--   rating_id    -> RatingId hex TEXT
--   rating_index -> INTEGER (position in the possible-ratings list, highest
--                            to lowest skill)
-- One rating per format: UNIQUE(format_id, rating_id) mirrors
-- FormatRating.UniquenessEquivalent.
CREATE TABLE format_rating (
    id           TEXT PRIMARY KEY,   -- RecordId hex string
    format_id    TEXT NOT NULL,      -- FormatId hex string
    rating_id    TEXT NOT NULL,      -- RatingId hex string
    rating_index INTEGER NOT NULL,   -- ordering within the format
    UNIQUE (format_id, rating_id)
);
