-- 0032_create_playoff_structures.sql
-- The playoff_structure table, matching the PlayoffStructure record shape
-- (model/playoff_structure.go). Table name equals record.Type() ("playoff_structure").
--   id              -> PlayoffStructureId hex TEXT primary key
--   user_id         -> UserId hex TEXT
--   byes            -> INTEGER
--   number_of_teams -> INTEGER
CREATE TABLE playoff_structure (
    id              TEXT PRIMARY KEY,   -- PlayoffStructureId hex string
    user_id         TEXT NOT NULL,      -- UserId hex string
    byes            INTEGER NOT NULL,
    number_of_teams INTEGER NOT NULL
);
