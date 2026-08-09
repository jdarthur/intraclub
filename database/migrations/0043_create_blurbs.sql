-- 0043_create_blurbs.sql
-- The blurb table, matching the Blurb record shape (model/blurb.go).
-- Table name equals record.Type() ("blurb").
--   id      -> BlurbId hex TEXT primary key
--   title   -> TEXT
--   content -> TEXT
--   owner   -> UserId hex TEXT
--   season  -> SeasonId hex TEXT
-- The former inline Blurb.Photos and Blurb.Reactions slices were normalized
-- into the blurb_photo (0044) and blurb_reaction (0046) child tables; they are
-- not part of this table.
CREATE TABLE blurb (
    id      TEXT PRIMARY KEY,   -- BlurbId hex string
    title   TEXT NOT NULL,
    content TEXT NOT NULL,
    owner   TEXT NOT NULL,      -- UserId hex string
    season  TEXT NOT NULL       -- SeasonId hex string
);
