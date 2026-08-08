-- 0026_create_format_lines.sql
-- The format_line join table, matching the FormatLine record shape
-- (model/format_line.go). This is the normalized replacement for the former
-- inline Format.Lines slice, preserving order via FormatIndex.
-- Table name equals record.Type() ("format_line").
--   id              -> RecordId hex TEXT primary key
--   format_id       -> FormatId hex TEXT
--   format_index    -> INTEGER (position of the line within the format)
--   player_1_rating -> RatingId hex TEXT
--   player_2_rating -> RatingId hex TEXT
-- One line per (format, index): UNIQUE(format_id, format_index) mirrors
-- FormatLine.UniquenessEquivalent.
CREATE TABLE format_line (
    id              TEXT PRIMARY KEY,   -- RecordId hex string
    format_id       TEXT NOT NULL,      -- FormatId hex string
    format_index    INTEGER NOT NULL,   -- line position within the format
    player_1_rating TEXT NOT NULL,      -- RatingId hex string
    player_2_rating TEXT NOT NULL,      -- RatingId hex string
    UNIQUE (format_id, format_index)
);
