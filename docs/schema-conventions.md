# SQLite Schema Conventions (Ticket #54)

This document defines the encoding and table conventions every SQLite migration
must follow. It is the authoritative reference for the model subtasks
(tickets #55–61) that add `CREATE TABLE` migrations on top of the skeleton
migration in `database/migrations/`.

## Column-type encoding conventions

Scalar values are encoded to columns as follows (see the reflection-based
mapping in `database/sqlite_db.go` for the runtime counterpart):

| Go type                          | Column type | Notes                                                                 |
| -------------------------------- | ----------- | --------------------------------------------------------------------- |
| `RecordId`, `UserId` (ID family) | `TEXT`      | 16-hex lowercase string (`RecordId.String()`), e.g. `0000000f7b1a3c2d` |
| `string`                         | `TEXT`      |                                                                       |
| `bool`                           | `INTEGER`   | `0` / `1`                                                             |
| `int` kinds (`int`, `int64`, …)  | `INTEGER`   |                                                                       |
| enum / named int                 | `INTEGER`   | stored as its integer value                                          |
| `time.Time`                      | `TEXT`      | RFC3339 with nanosecond precision (`time.RFC3339Nano`), UTC            |
| pointer                          | nullable    | `NULL` when nil                                                       |

### Hex IDs are `TEXT`, not `INTEGER`

ID-family types (`RecordId`, `UserId`, and any aliases) are stored as **hex
`TEXT`** columns, never as `INTEGER`:

- `RecordId` is a `uint64`; values above `MaxInt64` cannot fit in SQLite's
  signed 64-bit `INTEGER` without overflowing.
- Storing the canonical 16-hex string keeps the column a stable, sortable,
  lossless representation of `RecordId.String()` and round-trips exactly
  through `RecordIdFromString`.
- The primary key is the record's own hex string, so hex `TEXT` is the natural
  PK type as well.

## Primary key decision: `TEXT PRIMARY KEY`, not `INTEGER PRIMARY KEY`

Every record table uses the record's hex `RecordId` as its primary key:

```sql
CREATE TABLE example (
    id TEXT PRIMARY KEY,   -- RecordId hex string, e.g. 0000000f7b1a3c2d
    ...
);
```

- `INTEGER PRIMARY KEY` is SQLite's rowid alias and would impose its own
  auto-increment identity, conflicting with the application-owned
  `RecordId`/`NewRecordId()`.
- `TEXT PRIMARY KEY` makes the `id` column the application's id directly,
  so `INSERT`/`SELECT ... WHERE id = ?` work with `record.GetId().String()`
  with no conversion layer (see `SqliteDbProvider.Create` / `GetOne`).

The `id` column is always named `id`, is always handled via `GetId`/`SetId`
and is excluded from reflection-based field mapping (`database/sqlite_db.go`).

## Normalization: full normalization, not JSON columns

The schema uses **full normalization**:

- Every **map** and **slice of structs** on a model gets its own child/lookup
  table (a join table where appropriate), not a JSON column.
- The child table references its parent by the parent's hex `TEXT` id column
  and carries its own primary key.
- This matches the established join-table pattern used across the model layer
  (`SeasonTeam`, `FormatRating`, `TeamMatchIndividualMatch`,
  `WeeklyMatchupTeamMatchup`, `ScheduleMatchup`, `DraftRatingCutoff`,
  `CommissionerProposalVote`, …).

## Naming

- Table name = `record.Type()` (snake_case plural of the model, e.g. `users`,
  `weekly_matchups`).
- Non-id columns use the field's JSON tag (snake_case) as the column name;
  without a JSON tag, the snake_cased Go field name is used
  (`database/sqlite_db.go`, `columnName`).
- Foreign-key columns are named after the referenced id, e.g. `user_id`,
  `season_id`, `format_id`.

## Adding a migration

- Model subtasks only add a new file to `database/migrations/`, zero-padded and
  ordered after the skeleton (`0001_schema_skeleton.sql`), e.g.
  `0002_create_users.sql`, `0003_create_teams.sql`, ….
- Each migration is applied in order inside its own transaction and recorded in
  `schema_migrations` (see `database/migrate.go`). Re-running is a no-op.
- Never edit an already-applied migration; add a new one instead.
