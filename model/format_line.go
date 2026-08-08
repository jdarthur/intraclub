package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

// FormatLine is a join-table record that binds a Format to a Rating pairing
// (two RatingId values) that plays as a line in a matchup. It is the fully
// normalized replacement for the former inline `Format.Lines` slice (see
// model/format.go), following the same join-table pattern as FormatRating,
// TeamRating, DraftRatingCutoff, etc.
//
// FormatIndex preserves the ordering of lines within a format; it is the value
// referenced by `LineupPairing.FormatLineIndex`. A natural unique constraint
// on (FormatId, FormatIndex) prevents two lines from sharing a position.
type FormatLine struct {
	ID            database.RecordId `json:"id"`
	FormatId      FormatId          `json:"format_id"`
	FormatIndex   int               `json:"format_index"`
	Player1Rating RatingId          `json:"player_1_rating"`
	Player2Rating RatingId          `json:"player_2_rating"`
}

// EquivalentTo reports whether two lines pair the same two ratings, regardless
// of which player is listed first (i.e. 1/2 is equivalent to 2/1).
func (l *FormatLine) EquivalentTo(other FormatLine) bool {
	if l.Player1Rating == other.Player1Rating && l.Player2Rating == other.Player2Rating {
		return true
	}
	if l.Player1Rating == other.Player2Rating && l.Player2Rating == other.Player1Rating {
		return true
	}
	return false
}

func (l *FormatLine) String() string {
	return fmt.Sprintf("%s / %s", l.Player1Rating, l.Player2Rating)
}

func (l *FormatLine) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (l *FormatLine) SetOwner(userId database.UserId) {}

func (l *FormatLine) Type() string {
	return "format_line"
}

func (l *FormatLine) GetId() database.RecordId {
	return l.ID
}

func (l *FormatLine) SetId(id database.RecordId) {
	l.ID = id
}

func (l *FormatLine) StaticallyValid() error {
	return nil
}

func (l *FormatLine) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Format{}, l.FormatId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &Rating{}, l.Player1Rating.RecordId()); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &Rating{}, l.Player2Rating.RecordId())
}

func (l *FormatLine) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (l *FormatLine) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (l *FormatLine) NewRecord() database.CrudRecord {
	return new(FormatLine)
}

// UniquenessEquivalent enforces the natural unique constraint on
// (FormatId, FormatIndex): a format may only have one line at each position.
func (l *FormatLine) UniquenessEquivalent(other *FormatLine) error {
	if l.FormatId == other.FormatId && l.FormatIndex == other.FormatIndex {
		return fmt.Errorf("format %s already has a line at index %d", l.FormatId, l.FormatIndex)
	}
	return nil
}
