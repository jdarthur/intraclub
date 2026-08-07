# API Shape Decision Documentation

## WeeklyMatchupTeamMatchup Normalization (Ticket #39)

### Decision: Wire format reassembles relationship rows into the existing `matchups` shape

The `WeeklyMatchup` model no longer stores an inline `[]*TeamMatchup` slice. Instead, each matchup entry is normalized into a separate `WeeklyMatchupTeamMatchup` record.

For API consumers:
- **Serialization**: When returning a `WeeklyMatchup` via the API, callers should call `w.GetMatchups(ctx, db)` to reassemble the `[]*TeamMatchup` slice from the relationship table and include it in the response under the existing `matchups` JSON key.
- **Deserialization**: When receiving a `WeeklyMatchup` from the API, callers should create the `WeeklyMatchup` record first, then call `w.SetMatchups(ctx, db, matchups)` to persist the individual matchup entries.
- **Wire format**: The external JSON shape remains unchanged: `{ "id": "...", "week_id": "...", "season_id": "...", "matchups": [...] }`. The `matchups` field is virtual — it is computed on read and expanded on write.

### Why

This preserves backward compatibility with existing API consumers while achieving the normalization goal. The `TeamMatchup` value type remains as a convenience for serialization, and the `GetMatchups`/`SetMatchups` helpers provide a clean interface for callers.
