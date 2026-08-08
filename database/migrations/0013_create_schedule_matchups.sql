-- 0013_create_schedule_matchups.sql
-- The schedule_matchup join table, matching the ScheduleMatchup record shape
-- (model/schedule_matchup.go). This is the normalized replacement for the
-- former inline Schedule.Matchups []WeeklyMatchupId slice, preserving ordering
-- via the Position column.
-- Table name equals record.Type() ("schedule_matchup").
--   id                -> ScheduleMatchupId hex TEXT primary key
--   schedule_id       -> ScheduleId hex TEXT
--   weekly_matchup_id -> WeeklyMatchupId hex TEXT
--   position          -> INTEGER
-- A weekly matchup may only be assigned to a schedule once:
-- UNIQUE(schedule_id, weekly_matchup_id) mirrors ScheduleMatchup.UniquenessEquivalent.
CREATE TABLE schedule_matchup (
    id                TEXT PRIMARY KEY,   -- ScheduleMatchupId hex string
    schedule_id       TEXT NOT NULL,      -- ScheduleId hex string
    weekly_matchup_id TEXT NOT NULL,      -- WeeklyMatchupId hex string
    position          INTEGER NOT NULL,
    UNIQUE (schedule_id, weekly_matchup_id)
);
