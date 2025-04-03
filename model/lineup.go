package model

import (
	"fmt"
	"intraclub/common"
)

type LineupId common.RecordId

func (id LineupId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id LineupId) String() string {
	return id.RecordId().String()
}

type Lineup struct {
	ID     LineupId
	TeamId TeamId // TeamId for this particular Lineup
	WeekId WeekId // Week that this Lineup applies to
}

func (l *Lineup) UniquenessEquivalent(other *Lineup) error {
	if l.WeekId == other.WeekId && l.TeamId == other.TeamId {
		return fmt.Errorf("duplicate record for team ID & week ID")
	}
	return nil
}

func (l *Lineup) SetOwner(recordId common.RecordId) {
	// don't need to do anything here as ownership is enforced by
	// team captain or co-captain status
}

func (l *Lineup) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return EditableByTeamCaptainOrCoCaptains(db, l.TeamId)
}

func (l *Lineup) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return AccessibleByTeamMembers(db, l.TeamId)
}

func (l *Lineup) Type() string {
	return "lineup"
}

func (l *Lineup) GetId() common.RecordId {
	return l.ID.RecordId()
}

func (l *Lineup) SetId(id common.RecordId) {
	l.ID = LineupId(id)
}

func (l *Lineup) StaticallyValid() error {
	return nil
}

func (l *Lineup) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &Team{}, l.TeamId.RecordId())
	if err != nil {
		return err
	}
	err = common.ExistsById(db, &Week{}, l.WeekId.RecordId())
	if err != nil {
		return err
	}
	return nil
}

func (l *Lineup) GetFormat(db common.DatabaseProvider) (*Format, error) {
	week, err := common.GetExistingRecordById(db, &Week{}, l.WeekId.RecordId())
	if err != nil {
		return nil, err
	}
	draft, err := common.GetExistingRecordById(db, &Draft{}, week.DraftId.RecordId())
	if err != nil {
		return nil, err
	}
	return common.GetExistingRecordById(db, &Format{}, draft.Format.RecordId())
}
