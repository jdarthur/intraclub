package model

import (
	"context"
	"time"

	"intraclub/database"
)

// SeasonLateAddition is a join table record that links a Season to Users added after the draft.
// This allows tracking users who joined a season after the draft was completed.
type SeasonLateAddition struct {
	ID        database.RecordId `json:"id"`
	SeasonId  SeasonId          `json:"season_id"`
	UserId    database.UserId   `json:"user_id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as SeasonLateAddition has no specific owner.
func (s *SeasonLateAddition) GetOwner() database.UserId {
	return database.InvalidUserId
}

// SetOwner is a no-op as SeasonLateAddition has no specific owner.
func (s *SeasonLateAddition) SetOwner(userId database.UserId) {}

// AccessibleTo returns everyone as SeasonLateAddition records are public.
func (s *SeasonLateAddition) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (s *SeasonLateAddition) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

// Type returns the record type identifier for SeasonLateAddition.
func (s *SeasonLateAddition) Type() string {
	return "season_late_addition"
}

// GetId returns the unique identifier for this SeasonLateAddition record.
func (s *SeasonLateAddition) GetId() database.RecordId {
	return s.ID
}

// SetId sets the unique identifier for this SeasonLateAddition record.
func (s *SeasonLateAddition) SetId(id database.RecordId) {
	s.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for SeasonLateAddition.
func (s *SeasonLateAddition) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that both the referenced Season and User records exist.
func (s *SeasonLateAddition) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Season{}, s.SeasonId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &User{}, s.UserId.RecordId()); err != nil {
		return err
	}
	return nil
}

// Timestamps returns the create and update timestamps for this SeasonLateAddition record.
func (s *SeasonLateAddition) Timestamps() (time.Time, time.Time, *time.Time) {
	return s.CreatedAt, s.UpdatedAt, nil
}

// SetCreatedAt sets the creation timestamp for this SeasonLateAddition record.
func (s *SeasonLateAddition) SetCreatedAt(createdAt time.Time) {
	s.CreatedAt = createdAt
}

// SetUpdatedAt sets the update timestamp for this SeasonLateAddition record.
func (s *SeasonLateAddition) SetUpdatedAt(updatedAt time.Time) {
	s.UpdatedAt = updatedAt
}

func (s *SeasonLateAddition) NewRecord() database.CrudRecord {
	return new(SeasonLateAddition)
}
