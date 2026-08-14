-- 0053_add_week_closed.sql
-- Add the commissioner-closed flag to the week table. The Week model
-- (model/week.go) tracks this so a week is marked final once every team match
-- in it is complete. Stored as INTEGER (0/1), matching the provider's bool
-- mapping.
ALTER TABLE week ADD COLUMN closed INTEGER NOT NULL DEFAULT 0;
