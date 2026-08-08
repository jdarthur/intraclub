-- 0011_create_team_ratings.sql
-- The team_rating join table, matching the TeamRating record shape
-- (model/team.go). This is the normalized replacement for the former inline
-- Team.RatingsMap (user -> rating).
-- Table name equals record.Type() ("team_rating").
--   id        -> RecordId hex TEXT primary key
--   team_id   -> TeamId hex TEXT
--   user_id   -> UserId hex TEXT
--   rating_id -> RatingId hex TEXT
-- One rating per (team, user): UNIQUE(team_id, user_id) mirrors
-- TeamRating.UniquenessEquivalent.
CREATE TABLE team_rating (
    id        TEXT PRIMARY KEY,   -- RecordId hex string
    team_id   TEXT NOT NULL,      -- TeamId hex string
    user_id   TEXT NOT NULL,      -- UserId hex string
    rating_id TEXT NOT NULL,      -- RatingId hex string
    UNIQUE (team_id, user_id)
);
