-- 0034_create_ratings.sql
-- The rating table, matching the Rating record shape (model/rating.go).
-- Table name equals record.Type() ("rating").
--   id           -> RatingId hex TEXT primary key
--   user_id      -> UserId hex TEXT
--   name         -> TEXT
--   description  -> TEXT
-- One rating per name: UNIQUE(name) mirrors Rating.UniquenessEquivalent.
CREATE TABLE rating (
    id          TEXT PRIMARY KEY,   -- RatingId hex string
    user_id     TEXT NOT NULL,      -- UserId hex string
    name        TEXT NOT NULL,
    description TEXT NOT NULL,
    UNIQUE (name)
);
