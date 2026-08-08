-- 0005_create_seasons.sql
-- The season table, matching the Season record shape (model/season.go).
-- Table name equals record.Type() ("season").
--   id                 -> RecordId hex TEXT primary key
--   name               -> TEXT
--   facility           -> FacilityId hex TEXT (InvalidRecordId encodes as all-zero hex)
--   start_time         -> TEXT (StartTime is a defined type over time.Time; RFC3339)
--   draft_id           -> DraftId hex TEXT
--   schedule_id        -> ScheduleId hex TEXT
--   playoff_structure  -> PlayoffStructureId hex TEXT
--   owner              -> UserId hex TEXT
-- One season per draft: UNIQUE(draft_id) mirrors Season.UniquenessEquivalent.
CREATE TABLE season (
    id                TEXT PRIMARY KEY,  -- SeasonId hex string
    name              TEXT NOT NULL,
    facility          TEXT NOT NULL,     -- FacilityId hex string
    start_time        TEXT NOT NULL,     -- StartTime (RFC3339)
    draft_id          TEXT NOT NULL,     -- DraftId hex string
    schedule_id       TEXT NOT NULL,     -- ScheduleId hex string
    playoff_structure TEXT NOT NULL,     -- PlayoffStructureId hex string
    owner             TEXT NOT NULL,     -- UserId hex string
    UNIQUE (draft_id)
);
