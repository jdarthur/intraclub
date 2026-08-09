-- 0042_create_photos.sql
-- The photo table, matching the Photo record shape (model/photo.go).
-- Table name equals record.Type() ("photo").
--   id        -> PhotoId hex TEXT primary key
--   owner     -> UserId hex TEXT
--   alt_text  -> TEXT
--   contents  -> BLOB (the binary image payload, []byte)
--   file_type -> INTEGER (PhotoType)
CREATE TABLE photo (
    id        TEXT PRIMARY KEY,   -- PhotoId hex string
    owner     TEXT NOT NULL,      -- UserId hex string
    alt_text  TEXT NOT NULL,
    contents  BLOB NOT NULL,      -- binary image content
    file_type INTEGER NOT NULL    -- PhotoType
);
