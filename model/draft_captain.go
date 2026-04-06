package model

import (
	"intraclub/common"
	"time"
)

// DraftCaptain is a join table record that links a Draft to its team captains.
// Each record represents a Team-Captain assignment for a specific Draft.
type DraftCaptain struct {
	ID        common.RecordId `json:"id"`
	DraftId   DraftId         `json:"draft_id"`
	TeamId    TeamId          `json:"team_id"`
	CaptainId UserId          `json:"captain_id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as DraftCaptain has no specific owner.
func (d *DraftCaptain) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

// SetOwner is a no-op as DraftCaptain has no specific owner.
func (d *DraftCaptain) SetOwner(recordId common.RecordId) {}

// Type returns the record type identifier for DraftCaptain.
func (d *DraftCaptain) Type() string {
	return "draft_captain"
}

// GetId returns the unique identifier for this DraftCaptain record.
func (d *DraftCaptain) GetId() common.RecordId {
	return d.ID
}

// SetId sets the unique identifier for this DraftCaptain record.
func (d *DraftCaptain) SetId(id common.RecordId) {
	d.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for DraftCaptain.
func (d *DraftCaptain) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that the referenced Draft, Team, and User records all exist.
func (d *DraftCaptain) DynamicallyValid(db common.DatabaseProvider) error {
	if err := common.ExistsById(db, &Draft{}, d.DraftId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(db, &Team{}, d.TeamId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(db, &User{}, d.CaptainId.RecordId()); err != nil {
		return err
	}
	return nil
}

// AccessibleTo returns everyone as DraftCaptain records are public.
func (d *DraftCaptain) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

// EditableBy returns only sysadmins as only they can modify DraftCaptain records.
func (d *DraftCaptain) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.SysAdminRecordId}
}

// Timestamps returns the create and update timestamps for this DraftCaptain record.
func (d *DraftCaptain) Timestamps() (time.Time, time.Time, *time.Time) {
	return d.CreatedAt, d.UpdatedAt, nil
}

// SetCreatedAt sets the creation timestamp for this DraftCaptain record.
func (d *DraftCaptain) SetCreatedAt(createdAt time.Time) {
	d.CreatedAt = createdAt
}

// SetUpdatedAt sets the update timestamp for this DraftCaptain record.
func (d *DraftCaptain) SetUpdatedAt(updatedAt time.Time) {
	d.UpdatedAt = updatedAt
}
