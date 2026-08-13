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

func (id ScheduleId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id *ScheduleId) UnmarshalJSON(bytes []byte) error {
	rid := id.RecordId()
	if err := (*database.RecordId)(&rid).UnmarshalJSON(bytes); err != nil {
		return err
	}
	*id = ScheduleId(rid)
	return nil
}

type Schedule struct {
	ID       ScheduleId `json:"id"`
	SeasonId SeasonId   `json:"season_id"`
}

func (s *Schedule) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (s *Schedule) UniquenessEquivalent(other *Schedule) error {
	// can only have one schedule per season ID
	if s.SeasonId == other.SeasonId {
		return fmt.Errorf("duplicate schedule for season ID")
	}
	return nil
}

func (s *Schedule) SetOwner(userId database.UserId) {
	// don't need to do anything here as the ownership of the
	// Schedule record type is automatically inferred &
	// enforced by the associated Season assigned to it
}

func NewSchedule() *Schedule {
	return &Schedule{}
}

func (s *Schedule) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return EditableBySeason(ctx, db, s.SeasonId)
}

func (s *Schedule) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
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

func (s *Schedule) DynamicallyValid(ctx context.Context, db database.Provider) error {
	err := database.ExistsById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return err
	}

	// Validate each assigned weekly matchup. Matchups are separate records
	// (ScheduleMatchup) with a FK to this schedule, so they are created after
	// the Schedule record itself; skip the check when none are assigned yet.
	matchups, err := s.GetMatchups(ctx, db)
	if err != nil {
		return err
	}
	if len(matchups) == 0 {
		return nil
	}

	for _, m := range matchups {
		err = m.DynamicallyValid(ctx, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Schedule) PostCreate(ctx context.Context, db database.Provider) error {
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return err
	}
	season.ScheduleID = s.ID
	return database.UpdateOne(ctx, db, season)
}

func (s *Schedule) GetWeeks(ctx context.Context, db database.Provider) ([]*Week, error) {
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return nil, err
	}

	return database.GetAllWhere[*Week](ctx, db, func(_ context.Context, c *Week) bool {
		return c.DraftId == season.DraftId
	})
}

func (s *Schedule) IsScheduleComplete(ctx context.Context, db database.Provider) (bool, error) {
	weeks, err := s.GetWeeks(ctx, db)
	if err != nil {
		return false, err
	}
	matchups, err := s.GetMatchups(ctx, db)
	if err != nil {
		return false, err
	}
	return len(weeks) == len(matchups), nil
}

// GetMatchups retrieves all WeeklyMatchup records assigned to this Schedule by
// querying the ScheduleMatchup relationship table, sorted by Position.
func (s *Schedule) GetMatchups(ctx context.Context, db database.Provider) ([]*WeeklyMatchup, error) {
	records, err := database.GetAllWhere[*ScheduleMatchup](ctx, db, func(_ context.Context, sm *ScheduleMatchup) bool {
		return sm.ScheduleId == s.ID
	})
	if err != nil {
		return nil, err
	}

	// Sort by Position to preserve ordering
	for i := 0; i < len(records); i++ {
		for j := i + 1; j < len(records); j++ {
			if records[j].Position < records[i].Position {
				records[i], records[j] = records[j], records[i]
			}
		}
	}

	matchups := make([]*WeeklyMatchup, 0, len(records))
	for _, sm := range records {
		weeklyMatchup, err := database.GetExistingRecordById(ctx, db, &WeeklyMatchup{}, sm.WeeklyMatchupId.RecordId())
		if err != nil {
			return nil, err
		}
		matchups = append(matchups, weeklyMatchup)
	}
	return matchups, nil
}

// SetMatchups replaces all weekly matchups assigned to this Schedule.
// It deletes existing ScheduleMatchup records and creates new ones.
func (s *Schedule) SetMatchups(ctx context.Context, db database.Provider, matchups []WeeklyMatchupId) error {
	// Delete existing matchup records
	existing, err := database.GetAllWhere[*ScheduleMatchup](ctx, db, func(_ context.Context, sm *ScheduleMatchup) bool {
		return sm.ScheduleId == s.ID
	})
	if err != nil {
		return err
	}
	for _, sm := range existing {
		_, _, err = database.DeleteOneById(ctx, db, &ScheduleMatchup{}, sm.ID.RecordId())
		if err != nil {
			return err
		}
	}

	// Create new matchup records
	for i, m := range matchups {
		sm := NewScheduleMatchup(s.ID, m, i)
		_, err = database.CreateOne(ctx, db, sm)
		if err != nil {
			return err
		}
	}
	return nil
}

// PreDelete cascades deletion to all associated ScheduleMatchup records.
func (s *Schedule) PreDelete(ctx context.Context, db database.Provider) error {
	existing, err := database.GetAllWhere[*ScheduleMatchup](ctx, db, func(_ context.Context, sm *ScheduleMatchup) bool {
		return sm.ScheduleId == s.ID
	})
	if err != nil {
		return err
	}
	for _, sm := range existing {
		_, _, err = database.DeleteOneById(ctx, db, &ScheduleMatchup{}, sm.ID.RecordId())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Schedule) NewRecord() database.CrudRecord {
	return new(Schedule)
}
