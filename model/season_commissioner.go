package model

import (
	"context"
	"time"

	"intraclub/database"
)

// SeasonCommissioner is a join table record that links a Season to its Commissioner users.
// This allows multiple commissioners per season and supports tracking creation/modification timestamps.
type SeasonCommissioner struct {
	ID        database.RecordId `json:"id"`
	SeasonId  SeasonId          `json:"season_id"`
	UserId    UserId            `json:"user_id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as SeasonCommissioner has no specific owner.
func (s *SeasonCommissioner) GetOwner() database.RecordId {
	return database.InvalidRecordId
}

// SetOwner is a no-op as SeasonCommissioner has no specific owner.
func (s *SeasonCommissioner) SetOwner(recordId database.RecordId) {}

// AccessibleTo returns everyone as SeasonCommissioner records are public.
func (s *SeasonCommissioner) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return database.AccessibleToEveryone
}

func (s *SeasonCommissioner) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return []database.RecordId{database.SysAdminRecordId}
}

// Type returns the record type identifier for SeasonCommissioner.
func (s *SeasonCommissioner) Type() string {
	return "season_commissioner"
}

// GetId returns the unique identifier for this SeasonCommissioner record.
func (s *SeasonCommissioner) GetId() database.RecordId {
	return s.ID
}

// SetId sets the unique identifier for this SeasonCommissioner record.
func (s *SeasonCommissioner) SetId(id database.RecordId) {
	s.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for SeasonCommissioner.
func (s *SeasonCommissioner) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that both the referenced Season and User records exist.
func (s *SeasonCommissioner) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	if err := database.ExistsById(ctx, db, &Season{}, s.SeasonId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &User{}, s.UserId.RecordId()); err != nil {
		return err
	}
	return nil
}

// Timestamps returns the create and update timestamps for this SeasonCommissioner record.
func (s *SeasonCommissioner) Timestamps() (time.Time, time.Time, *time.Time) {
	return s.CreatedAt, s.UpdatedAt, nil
}

// SetCreatedAt sets the creation timestamp for this SeasonCommissioner record.
func (s *SeasonCommissioner) SetCreatedAt(createdAt time.Time) {
	s.CreatedAt = createdAt
}

// SetUpdatedAt sets the update timestamp for this SeasonCommissioner record.
func (s *SeasonCommissioner) SetUpdatedAt(updatedAt time.Time) {
	s.UpdatedAt = updatedAt
}

func (s *SeasonCommissioner) BlankRecord() database.CrudRecord {
	return new(SeasonCommissioner)
}
