-- 0012_create_schedules.sql
-- The schedule table, matching the Schedule record shape (model/schedule.go).
-- Table name equals record.Type() ("schedule").
--   id        -> ScheduleId hex TEXT primary key
--   season_id -> SeasonId hex TEXT
-- One schedule per season: UNIQUE(season_id) mirrors Schedule.UniquenessEquivalent.
-- Schedule.Matchups ([]WeeklyMatchupId) is normalized into the schedule_matchup
-- join table (0013), not stored inline.
CREATE TABLE schedule (
    id        TEXT PRIMARY KEY,   -- ScheduleId hex string
    season_id TEXT NOT NULL       -- SeasonId hex string
);
