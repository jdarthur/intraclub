package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

type LineupPairingId database.RecordId

func (id LineupPairingId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id LineupPairingId) String() string {
	return id.RecordId().String()
}

type LineupPairing struct {
	ID              LineupPairingId // unique ID for this LineupPairing
	LineupId        LineupId        // This correlates a LineupPairing into a group with other pairing and assigns to a Week
	TeamId          TeamId          // Players must be on this team
	Player1         database.UserId          // Player in slot 1 for the format / line
	Player2         database.UserId          // Player in slot 2 for the format / line
	FormatLineIndex int             // index in the Format.Lines list that this pairing applies to
}

func (l *LineupPairing) UniquenessEquivalent(other *LineupPairing) error {
	if l.LineupId == other.LineupId {
		if l.Player1 == other.Player1 {
			return fmt.Errorf("duplicate record for lineup %s and player 1 %s", l.LineupId, other.Player1)
		}
		if l.Player2 == other.Player2 {
			return fmt.Errorf("duplicate record for lineup %s and player 2 %s", l.LineupId, other.Player2)
		}
	}
	return nil
}

func (l *LineupPairing) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (l *LineupPairing) Type() string {
	return "lineup_pairing"
}

func (l *LineupPairing) GetId() database.RecordId {
	return l.ID.RecordId()
}

func (l *LineupPairing) SetId(id database.RecordId) {
	l.ID = LineupPairingId(id)
}

func (l *LineupPairing) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return EditableByTeamCaptainOrCoCaptains(ctx, db, l.TeamId)
}

func (l *LineupPairing) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return AccessibleByTeamMembers(ctx, db, l.TeamId)
}

func (l *LineupPairing) SetOwner(userId database.UserId) {
	// don't need to do anything here as editable-by rights
	// are enforced via the team ID
}

func (l *LineupPairing) StaticallyValid() error {
	if l.Player1 == l.Player2 {
		return fmt.Errorf("player 1 ID is the same as player 2 ID")
	}
	return nil
}

func (l *LineupPairing) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {

	// get team
	team, err := database.GetExistingRecordById(ctx, db, &Team{}, l.TeamId.RecordId())
	if err != nil {
		return err
	}

	// validate that both players are members of this team
	isMember1, err := team.IsTeamMember(ctx, db, l.Player1)
	if err != nil {
		return err
	}
	if !isMember1 {
		return fmt.Errorf("player 1 is not team member for team %s", l.TeamId)
	}
	isMember2, err := team.IsTeamMember(ctx, db, l.Player2)
	if err != nil {
		return err
	}
	if !isMember2 {
		return fmt.Errorf("player 2 is not team member for team %s", l.TeamId)
	}

	format, err := l.GetFormat(ctx, db)
	if err != nil {
		return err
	}

	// check that the line index for this pairing is within the bounds of the format
	if l.FormatLineIndex >= len(format.Lines) {
		return fmt.Errorf("format line index out of range: %d, max %d", l.FormatLineIndex, len(format.Lines)-1)
	}
	return nil
}

func (l *LineupPairing) GetFormat(ctx context.Context, db database.DatabaseProvider) (*Format, error) {
	// get lineup so that we can get the format
	lineup, err := database.GetExistingRecordById(ctx, db, &Lineup{}, l.LineupId.RecordId())
	if err != nil {
		return nil, err
	}

	// get the format for the lineup to validate the correctness of the line
	// index and each players' ratings
	return lineup.GetFormat(ctx, db)
}

func (l *LineupPairing) ValidatePlayerRatings(ctx context.Context, db database.DatabaseProvider) error {
	format, err := l.GetFormat(ctx, db)
	if err != nil {
		return err
	}
	team, err := database.GetExistingRecordById(ctx, db, &Team{}, l.TeamId.RecordId())
	if err != nil {
		return err
	}
	return l._validatePlayerRatings(format, team)
}

func (l *LineupPairing) _validatePlayerRatings(format *Format, team *Team) error {
	line := format.Lines[l.FormatLineIndex]

	rating1 := team.RatingsMap[l.Player1]
	if rating1 != line.Player1Rating {
		return fmt.Errorf("player 1 has rating %s, expected %s for line index %d for format", line.Player1Rating, rating1, l.FormatLineIndex)
	}

	rating2 := team.RatingsMap[l.Player2]
	if rating2 != line.Player2Rating {
		return fmt.Errorf("player 2 has rating %s, expected %s for line index %d for format", line.Player2Rating, rating2, l.FormatLineIndex)
	}
	return nil
}

func (l *LineupPairing) BlankRecord() database.CrudRecord {
	return new(LineupPairing)
}
