package model

import (
	"errors"
	"fmt"
	"intraclub/common"
)

type lineupPairing struct {
	Format  common.RecordId // ID of a particular Format for a Season
	Line    int             // index of the line inside the Format
	Player1 UserId          // index of the User who is player one for this Format / Line combination
	Player2 UserId          // index of the User who is player two for this Format / Line combination
}

func (l lineupPairing) StaticallyValid() error {
	if l.Line < 0 {
		return errors.New("line is less than zero")
	}
	return nil
}

func (l lineupPairing) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &User{}, l.Player1.RecordId())
	if err != nil {
		return err
	}

	err = common.ExistsById(db, &User{}, l.Player2.RecordId())
	if err != nil {
		return err
	}

	format, exists, err := common.GetOneById(db, &Format{}, l.Format)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("format %s does not exist", l.Format)
	}

	if l.Line > len(format.Lines) {
		return fmt.Errorf("line %d is greater than number of lines in format (%d)", l.Line, len(format.Lines))
	}

	return nil
}

type Lineup struct {
	ID       common.RecordId
	TeamId   TeamId
	Pairings []lineupPairing
}

func (l Lineup) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return EditableByTeamCaptainOrCoCaptains(db, l.TeamId)
}

func (l Lineup) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return AccessibleByTeamMembers(db, l.TeamId)
}

func (l Lineup) Type() string {
	return "lineup"
}

func (l Lineup) GetId() common.RecordId {
	return l.ID
}

func (l Lineup) SetId(id common.RecordId) {
	l.ID = id
}

func (l Lineup) StaticallyValid() error {
	return nil
}

func (l Lineup) DynamicallyValid(db common.DatabaseProvider) error {
	for _, pairing := range l.Pairings {
		err := pairing.StaticallyValid()
		if err != nil {
			return err
		}
		err = pairing.DynamicallyValid(db)
		if err != nil {
			return err
		}
	}
	return nil
}
