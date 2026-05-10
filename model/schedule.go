package model

import (
	"context"
	"fmt"
	"intraclub/common"
)

type ScheduleId common.RecordId

func (id ScheduleId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id ScheduleId) String() string {
	return id.RecordId().String()
}

type Schedule struct {
	ID       ScheduleId
	SeasonId SeasonId
	Matchups []WeeklyMatchupId
}

func (s *Schedule) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

func (s *Schedule) UniquenessEquivalent(other *Schedule) error {
	// can only have one schedule per season ID
	if s.SeasonId == other.SeasonId {
		return fmt.Errorf("duplicate schedule for season ID")
	}
	return nil
}

func (s *Schedule) SetOwner(recordId common.RecordId) {
	// don't need to do anything here as the ownership of the
	// Schedule record type is automatically inferred &
	// enforced by the associated Season assigned to it
}

func NewSchedule() *Schedule {
	return &Schedule{}
}

func (s *Schedule) EditableBy(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	return EditableBySeason(ctx, db, s.SeasonId)
}

func (s *Schedule) AccessibleTo(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (s *Schedule) Type() string {
	return "schedule"
}

func (s *Schedule) GetId() common.RecordId {
	return s.ID.RecordId()
}

func (s *Schedule) SetId(id common.RecordId) {
	s.ID = ScheduleId(id)
}

func (s *Schedule) StaticallyValid() error {
	return nil
}

func (s *Schedule) DynamicallyValid(ctx context.Context, db common.DatabaseProvider) error {
	err := common.ExistsById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return err
	}

	for _, m := range s.Matchups {
		weeklyMatchup, err := common.GetExistingRecordById(ctx, db, &WeeklyMatchup{}, m.RecordId())
		if err != nil {
			return err
		}
		err = weeklyMatchup.DynamicallyValid(ctx, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Schedule) PostCreate(ctx context.Context, db common.DatabaseProvider) error {
	season, err := common.GetExistingRecordById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return err
	}
	season.ScheduleID = s.ID
	return common.UpdateOne(ctx, db, season)
}

func (s *Schedule) GetWeeks(ctx context.Context, db common.DatabaseProvider) ([]*Week, error) {
	season, err := common.GetExistingRecordById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return nil, err
	}

	return common.GetAllWhere[*Week](ctx, db, func(_ context.Context, c *Week) bool {
		return c.DraftId == season.DraftId
	})
}

func (s *Schedule) IsScheduleComplete(ctx context.Context, db common.DatabaseProvider) (bool, error) {
	weeks, err := s.GetWeeks(ctx, db)
	if err != nil {
		return false, err
	}
	return len(weeks) == len(s.Matchups), nil
}

func (s *Schedule) BlankRecord() common.CrudRecord {
	return new(Schedule)
}
