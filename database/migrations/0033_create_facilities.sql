-- 0033_create_facilities.sql
-- The facility table, matching the Facility record shape (model/facility.go).
-- Table name equals record.Type() ("facility").
--   id            -> FacilityId hex TEXT primary key
--   user_id       -> UserId hex TEXT
--   name          -> TEXT
--   address       -> TEXT
--   courts        -> INTEGER (Facility.NumberOfCourts)
--   layout_photo  -> PhotoId hex TEXT
CREATE TABLE facility (
    id            TEXT PRIMARY KEY,   -- FacilityId hex string
    user_id       TEXT NOT NULL,      -- UserId hex string
    name          TEXT NOT NULL,
    address       TEXT NOT NULL,
    courts        INTEGER NOT NULL,
    layout_photo  TEXT NOT NULL       -- PhotoId hex string
);
