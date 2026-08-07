package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

// ScheduleMatchupId is the unique identifier for a ScheduleMatchup record.
type ScheduleMatchupId database.RecordId

func (id ScheduleMatchupId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id ScheduleMatchupId) String() string {
	return id.RecordId().String()
}

// ScheduleMatchup is a join table record that assigns a WeeklyMatchup to a
// Schedule at a given Position (preserving ordering). This replaces the former
// denormalized `Schedule.Matchups` inline ID-slice, enabling the relationship
// to be queried/indexed individually and stored in its own collection/table.
//
// It is stored in its own collection (`schedule_matchup`) with FKs to schedule
// and weekly_matchup, and a natural unique constraint on
// (ScheduleId, WeeklyMatchupId). When the #36 SQLite provider lands, this
// becomes a `schedule_matchups` table with a migration/backfill.
type ScheduleMatchup struct {
	ID              ScheduleMatchupId `json:"id"`
	ScheduleId      ScheduleId        `json:"schedule_id"`
	WeeklyMatchupId WeeklyMatchupId   `json:"weekly_matchup_id"`
	Position        int               `json:"position"`
}

// GetOwner returns InvalidUserId as ScheduleMatchup has no specific owner.
func (s *ScheduleMatchup) GetOwner() database.UserId {
	return database.InvalidUserId
}

// SetOwner is a no-op as ScheduleMatchup has no specific owner.
func (s *ScheduleMatchup) SetOwner(userId database.UserId) {}

// AccessibleTo returns everyone as ScheduleMatchup records are public.
func (s *ScheduleMatchup) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

// EditableBy returns sysadmin-only access for join records.
func (s *ScheduleMatchup) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

// Type returns the record type identifier for ScheduleMatchup.
func (s *ScheduleMatchup) Type() string {
	return "schedule_matchup"
}

// GetId returns the unique identifier for this record.
func (s *ScheduleMatchup) GetId() database.RecordId {
	return s.ID.RecordId()
}

// SetId sets the unique identifier for this record.
func (s *ScheduleMatchup) SetId(id database.RecordId) {
	s.ID = ScheduleMatchupId(id)
}

// StaticallyValid always returns nil as there are no static validation rules.
func (s *ScheduleMatchup) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that the referenced Schedule and WeeklyMatchup
// records exist.
func (s *ScheduleMatchup) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Schedule{}, s.ScheduleId.RecordId()); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &WeeklyMatchup{}, s.WeeklyMatchupId.RecordId())
}

// UniquenessEquivalent enforces the natural unique constraint on
// (ScheduleId, WeeklyMatchupId): a weekly matchup may only be assigned to a
// schedule once.
func (s *ScheduleMatchup) UniquenessEquivalent(other *ScheduleMatchup) error {
	if s.ScheduleId == other.ScheduleId && s.WeeklyMatchupId == other.WeeklyMatchupId {
		return fmt.Errorf("schedule %s already has weekly matchup %s assigned", s.ScheduleId, s.WeeklyMatchupId)
	}
	return nil
}

// NewRecord returns a new empty ScheduleMatchup.
func (s *ScheduleMatchup) NewRecord() database.CrudRecord {
	return new(ScheduleMatchup)
}

// NewScheduleMatchup creates a new ScheduleMatchup record.
func NewScheduleMatchup(scheduleId ScheduleId, weeklyMatchupId WeeklyMatchupId, position int) *ScheduleMatchup {
	return &ScheduleMatchup{
		ScheduleId:      scheduleId,
		WeeklyMatchupId: weeklyMatchupId,
		Position:        position,
	}
}
