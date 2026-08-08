-- 0024_create_formats.sql
-- The format table, matching the Format record shape (model/format.go).
-- Table name equals record.Type() ("format").
--   id      -> FormatId hex TEXT primary key
--   user_id -> UserId hex TEXT
--   name    -> TEXT
--
-- Format.PossibleRatings was normalized into the format_rating join table
-- (0025) and Format.Lines into the format_line join table (0026); neither
-- inline slice is part of this table.
CREATE TABLE format (
    id      TEXT PRIMARY KEY,   -- FormatId hex string
    user_id TEXT NOT NULL,      -- UserId hex string
    name    TEXT NOT NULL
);
