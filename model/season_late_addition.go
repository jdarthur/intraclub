package model

import (
	"intraclub/common"
	"time"
)

// SeasonLateAddition is a join table record that links a Season to Users added after the draft.
// This allows tracking users who joined a season after the draft was completed.
type SeasonLateAddition struct {
	ID        common.RecordId `json:"id"`
	SeasonId  SeasonId        `json:"season_id"`
	UserId    UserId          `json:"user_id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as SeasonLateAddition has no specific owner.
func (s *SeasonLateAddition) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

// SetOwner is a no-op as SeasonLateAddition has no specific owner.
func (s *SeasonLateAddition) SetOwner(recordId common.RecordId) {}

// AccessibleTo returns everyone as SeasonLateAddition records are public.
func (s *SeasonLateAddition) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

// EditableBy returns only sysadmins as only they can modify SeasonLateAddition records.
func (s *SeasonLateAddition) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.SysAdminRecordId}
}

// Type returns the record type identifier for SeasonLateAddition.
func (s *SeasonLateAddition) Type() string {
	return "season_late_addition"
}

// GetId returns the unique identifier for this SeasonLateAddition record.
func (s *SeasonLateAddition) GetId() common.RecordId {
	return s.ID
}

// SetId sets the unique identifier for this SeasonLateAddition record.
func (s *SeasonLateAddition) SetId(id common.RecordId) {
	s.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for SeasonLateAddition.
func (s *SeasonLateAddition) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that both the referenced Season and User records exist.
func (s *SeasonLateAddition) DynamicallyValid(db common.DatabaseProvider) error {
	if err := common.ExistsById(db, &Season{}, s.SeasonId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(db, &User{}, s.UserId.RecordId()); err != nil {
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

func (s *SeasonLateAddition) BlankRecord() common.CrudRecord {
	return new(SeasonLateAddition)
}
