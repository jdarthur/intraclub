package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

type LineupId database.RecordId

func (id LineupId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id LineupId) String() string {
	return id.RecordId().String()
}

type Lineup struct {
	ID     LineupId
	TeamId TeamId // TeamId for this particular Lineup
	WeekId WeekId // Week that this Lineup applies to
}

func (l *Lineup) GetOwner() database.RecordId {
	return database.InvalidRecordId
}

func (l *Lineup) UniquenessEquivalent(other *Lineup) error {
	if l.WeekId == other.WeekId && l.TeamId == other.TeamId {
		return fmt.Errorf("duplicate record for team ID & week ID")
	}
	return nil
}

func (l *Lineup) SetOwner(recordId database.RecordId) {
	// don't need to do anything here as ownership is enforced by
	// team captain or co-captain status
}

func (l *Lineup) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return EditableByTeamCaptainOrCoCaptains(ctx, db, l.TeamId)
}

func (l *Lineup) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return AccessibleByTeamMembers(ctx, db, l.TeamId)
}

func (l *Lineup) Type() string {
	return "lineup"
}

func (l *Lineup) GetId() database.RecordId {
	return l.ID.RecordId()
}

func (l *Lineup) SetId(id database.RecordId) {
	l.ID = LineupId(id)
}

func (l *Lineup) StaticallyValid() error {
	return nil
}

func (l *Lineup) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	err := database.ExistsById(ctx, db, &Team{}, l.TeamId.RecordId())
	if err != nil {
		return err
	}
	err = database.ExistsById(ctx, db, &Week{}, l.WeekId.RecordId())
	if err != nil {
		return err
	}
	return nil
}

func (l *Lineup) GetFormat(ctx context.Context, db database.DatabaseProvider) (*Format, error) {
	week, err := database.GetExistingRecordById(ctx, db, &Week{}, l.WeekId.RecordId())
	if err != nil {
		return nil, err
	}
	draft, err := database.GetExistingRecordById(ctx, db, &Draft{}, week.DraftId.RecordId())
	if err != nil {
		return nil, err
	}
	return database.GetExistingRecordById(ctx, db, &Format{}, draft.Format.RecordId())
}

func (l *Lineup) BlankRecord() database.CrudRecord {
	return new(Lineup)
}
