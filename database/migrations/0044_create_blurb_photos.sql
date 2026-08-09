-- 0044_create_blurb_photos.sql
-- The blurb_photo join table, matching the BlurbPhoto record shape
-- (model/blurb.go). This is the normalized replacement for the former inline
-- Blurb.Photos slice.
-- Table name equals record.Type() ("blurb_photo").
--   id       -> RecordId hex TEXT primary key
--   blurb_id -> BlurbId hex TEXT
--   photo_id -> PhotoId hex TEXT
CREATE TABLE blurb_photo (
    id       TEXT PRIMARY KEY,   -- RecordId hex string
    blurb_id TEXT NOT NULL,      -- BlurbId hex string
    photo_id TEXT NOT NULL       -- PhotoId hex string
);
