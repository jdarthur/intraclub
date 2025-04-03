package model

import (
	"fmt"
	"intraclub/common"
)

type LineupPairingId common.RecordId

func (id LineupPairingId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id LineupPairingId) String() string {
	return id.RecordId().String()
}

type LineupPairing struct {
	ID              LineupPairingId // unique ID for this LineupPairing
	LineupId        LineupId        // This correlates a LineupPairing into a group with other pairing and assigns to a Week
	TeamId          TeamId          // Players must be on this team
	Player1         UserId          // Player in slot 1 for the format / line
	Player2         UserId          // Player in slot 2 for the format / line
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

func (l *LineupPairing) Type() string {
	return "lineup_pairing"
}

func (l *LineupPairing) GetId() common.RecordId {
	return l.ID.RecordId()
}

func (l *LineupPairing) SetId(id common.RecordId) {
	l.ID = LineupPairingId(id)
}

func (l *LineupPairing) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return EditableByTeamCaptainOrCoCaptains(db, l.TeamId)
}

func (l *LineupPairing) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return AccessibleByTeamMembers(db, l.TeamId)
}

func (l *LineupPairing) SetOwner(recordId common.RecordId) {
	// don't need to do anything here as editable-by rights
	// are enforced via the team ID
}

func (l *LineupPairing) StaticallyValid() error {
	if l.Player1 == l.Player2 {
		return fmt.Errorf("player 1 ID is the same as player 2 ID")
	}
	return nil
}

func (l *LineupPairing) DynamicallyValid(db common.DatabaseProvider) error {

	// get team
	team, err := common.GetExistingRecordById(db, &Team{}, l.TeamId.RecordId())
	if err != nil {
		return err
	}

	// validate that both players are members of this team
	if !team.IsTeamMember(l.Player1) {
		return fmt.Errorf("player 1 is not team member for team %s", l.TeamId)
	}
	if !team.IsTeamMember(l.Player2) {
		return fmt.Errorf("player 2 is not team member for team %s", l.TeamId)
	}

	format, err := l.GetFormat(db)
	if err != nil {
		return err
	}

	// check that the line index for this pairing is within the bounds of the format
	if l.FormatLineIndex >= len(format.Lines) {
		return fmt.Errorf("format line index out of range: %d, max %d", l.FormatLineIndex, len(format.Lines)-1)
	}
	return nil
}

func (l *LineupPairing) GetFormat(db common.DatabaseProvider) (*Format, error) {
	// get lineup so that we can get the format
	lineup, err := common.GetExistingRecordById(db, &Lineup{}, l.LineupId.RecordId())
	if err != nil {
		return nil, err
	}

	// get the format for the lineup to validate the correctness of the line
	// index and each players' ratings
	return lineup.GetFormat(db)
}

func (l *LineupPairing) ValidatePlayerRatings(db common.DatabaseProvider) error {
	format, err := l.GetFormat(db)
	if err != nil {
		return err
	}
	team, err := common.GetExistingRecordById(db, &Team{}, l.TeamId.RecordId())
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
