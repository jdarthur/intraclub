-- 0023_create_pre_draft_grades.sql
-- The pre_draft_grade table, matching the PreDraftGrade record shape
-- (model/pre_draft_grade.go).
-- Table name equals record.Type() ("pre_draft_grade").
--   id        -> RecordId hex TEXT primary key
--   player_id -> UserId hex TEXT
--   draft_id  -> DraftId hex TEXT
--   grader_id -> UserId hex TEXT
--   modifier  -> INTEGER (PreDraftRatingModifier: 0 weak, 1 average, 2 strong)
--   rating    -> RatingId hex TEXT
-- One grade per (player, grader, draft): UNIQUE(player_id, grader_id, draft_id)
-- mirrors PreDraftGrade.UniquenessEquivalent.
CREATE TABLE pre_draft_grade (
    id        TEXT PRIMARY KEY,   -- RecordId hex string
    player_id TEXT NOT NULL,      -- UserId hex string
    draft_id  TEXT NOT NULL,      -- DraftId hex string
    grader_id TEXT NOT NULL,      -- UserId hex string
    modifier  INTEGER NOT NULL,
    rating    TEXT NOT NULL,      -- RatingId hex string
    UNIQUE (player_id, grader_id, draft_id)
);
