# API Shape Decision Documentation

## ScheduleMatchup Normalization (Ticket #38)

### Decision: Wire format reassembles relationship rows into the existing `matchups` shape

The `Schedule` model no longer stores an inline `[]WeeklyMatchupId` slice. Instead, each weekly matchup assignment is normalized into a separate `ScheduleMatchup` record (`schedule_matchup` collection/table) with FKs to `schedule` and `weekly_matchup`, a natural unique constraint on `(ScheduleId, WeeklyMatchupId)`, and a `Position` to preserve ordering.

For API consumers:
- **Serialization**: When returning a `Schedule` via the API, callers should call `s.GetMatchups(ctx, db)` to reassemble the `[]*WeeklyMatchup` slice from the relationship table and include it in the response under the existing `matchups` JSON key.
- **Deserialization**: When receiving a `Schedule` from the API, callers should create the `Schedule` record first, then call `s.SetMatchups(ctx, db, ids)` to persist the individual matchup assignments.
- **Wire format**: The external JSON shape remains unchanged: `{ "id": "...", "season_id": "...", "matchups": [...] }`. The `matchups` field is virtual — it is computed on read and expanded on write.

### Why

This preserves backward compatibility with existing API consumers while achieving the normalization goal. The `GetMatchups`/`SetMatchups` helpers provide a clean interface for callers, and the `Position` column preserves the ordering that the former inline list implied.

> Note: The `schedule_matchups` table creation + backfill from the former inline data is deferred to the #36 SQLite provider migration framework, as documented in the `ScheduleMatchup` type comment.
