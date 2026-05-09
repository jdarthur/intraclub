package model

import (
	"intraclub/common"
	"time"
)

// SeasonTeam is a join table record that links a Season to its participating Teams.
// This allows multiple teams per season and supports tracking creation/modification timestamps.
type SeasonTeam struct {
	ID        common.RecordId `json:"id"`
	SeasonId  SeasonId        `json:"season_id"`
	TeamId    TeamId          `json:"team_id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as SeasonTeam has no specific owner.
func (s *SeasonTeam) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

// SetOwner is a no-op as SeasonTeam has no specific owner.
func (s *SeasonTeam) SetOwner(recordId common.RecordId) {}

// AccessibleTo returns everyone as SeasonTeam records are public.
func (s *SeasonTeam) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

// EditableBy returns only sysadmins as only they can modify SeasonTeam records.
func (s *SeasonTeam) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.SysAdminRecordId}
}

// Type returns the record type identifier for SeasonTeam.
func (s *SeasonTeam) Type() string {
	return "season_team"
}

// GetId returns the unique identifier for this SeasonTeam record.
func (s *SeasonTeam) GetId() common.RecordId {
	return s.ID
}

// SetId sets the unique identifier for this SeasonTeam record.
func (s *SeasonTeam) SetId(id common.RecordId) {
	s.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for SeasonTeam.
func (s *SeasonTeam) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that both the referenced Season and Team records exist.
func (s *SeasonTeam) DynamicallyValid(db common.DatabaseProvider) error {
	if err := common.ExistsById(db, &Season{}, s.SeasonId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(db, &Team{}, s.TeamId.RecordId()); err != nil {
		return err
	}
	return nil
}

// Timestamps returns the create and update timestamps for this SeasonTeam record.
func (s *SeasonTeam) Timestamps() (time.Time, time.Time, *time.Time) {
	return s.CreatedAt, s.UpdatedAt, nil
}

// SetCreatedAt sets the creation timestamp for this SeasonTeam record.
func (s *SeasonTeam) SetCreatedAt(createdAt time.Time) {
	s.CreatedAt = createdAt
}

// SetUpdatedAt sets the update timestamp for this SeasonTeam record.
func (s *SeasonTeam) SetUpdatedAt(updatedAt time.Time) {
	s.UpdatedAt = updatedAt
}

func (s *SeasonTeam) BlankRecord() common.CrudRecord {
	return new(SeasonTeam)
}
