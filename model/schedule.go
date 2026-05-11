package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

type ScheduleId database.RecordId

func (id ScheduleId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id ScheduleId) String() string {
	return id.RecordId().String()
}

type Schedule struct {
	ID       ScheduleId
	SeasonId SeasonId
	Matchups []WeeklyMatchupId
}

func (s *Schedule) GetOwner() database.RecordId {
	return database.InvalidRecordId
}

func (s *Schedule) UniquenessEquivalent(other *Schedule) error {
	// can only have one schedule per season ID
	if s.SeasonId == other.SeasonId {
		return fmt.Errorf("duplicate schedule for season ID")
	}
	return nil
}

func (s *Schedule) SetOwner(recordId database.RecordId) {
	// don't need to do anything here as the ownership of the
	// Schedule record type is automatically inferred &
	// enforced by the associated Season assigned to it
}

func NewSchedule() *Schedule {
	return &Schedule{}
}

func (s *Schedule) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return EditableBySeason(ctx, db, s.SeasonId)
}

func (s *Schedule) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return database.AccessibleToEveryone
}

func (s *Schedule) Type() string {
	return "schedule"
}

func (s *Schedule) GetId() database.RecordId {
	return s.ID.RecordId()
}

func (s *Schedule) SetId(id database.RecordId) {
	s.ID = ScheduleId(id)
}

func (s *Schedule) StaticallyValid() error {
	return nil
}

func (s *Schedule) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	err := database.ExistsById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return err
	}

	for _, m := range s.Matchups {
		weeklyMatchup, err := database.GetExistingRecordById(ctx, db, &WeeklyMatchup{}, m.RecordId())
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

func (s *Schedule) PostCreate(ctx context.Context, db database.DatabaseProvider) error {
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return err
	}
	season.ScheduleID = s.ID
	return database.UpdateOne(ctx, db, season)
}

func (s *Schedule) GetWeeks(ctx context.Context, db database.DatabaseProvider) ([]*Week, error) {
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return nil, err
	}

	return database.GetAllWhere[*Week](ctx, db, func(_ context.Context, c *Week) bool {
		return c.DraftId == season.DraftId
	})
}

func (s *Schedule) IsScheduleComplete(ctx context.Context, db database.DatabaseProvider) (bool, error) {
	weeks, err := s.GetWeeks(ctx, db)
	if err != nil {
		return false, err
	}
	return len(weeks) == len(s.Matchups), nil
}

func (s *Schedule) BlankRecord() database.CrudRecord {
	return new(Schedule)
}
