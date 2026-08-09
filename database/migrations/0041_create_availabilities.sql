-- 0041_create_availabilities.sql
-- The availability table, matching the Availability record shape
-- (model/availability.go).
-- Table name equals record.Type() ("availability").
--   id        -> RecordId hex TEXT primary key
--   user_id   -> UserId hex TEXT
--   week_id   -> WeekId hex TEXT
--   available -> INTEGER (AvailabilityOption)
-- One availability per (user, week): UNIQUE(user_id, week_id) mirrors
-- Availability.UniquenessEquivalent.
CREATE TABLE availability (
    id        TEXT PRIMARY KEY,   -- RecordId hex string
    user_id   TEXT NOT NULL,      -- UserId hex string
    week_id   TEXT NOT NULL,      -- WeekId hex string
    available INTEGER NOT NULL,
    UNIQUE (user_id, week_id)
);
