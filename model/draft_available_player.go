package model

import (
	"intraclub/common"
	"time"
)

// DraftAvailablePlayer is a join table record that links a Draft to available players.
// Each record represents a User who is eligible to be drafted in a specific Draft.
type DraftAvailablePlayer struct {
	ID        common.RecordId `json:"id"`
	DraftId   DraftId         `json:"draft_id"`
	PlayerId  UserId          `json:"player_id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as DraftAvailablePlayer has no specific owner.
func (d *DraftAvailablePlayer) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

// SetOwner is a no-op as DraftAvailablePlayer has no specific owner.
func (d *DraftAvailablePlayer) SetOwner(recordId common.RecordId) {}

// Type returns the record type identifier for DraftAvailablePlayer.
func (d *DraftAvailablePlayer) Type() string {
	return "draft_available_player"
}

// GetId returns the unique identifier for this DraftAvailablePlayer record.
func (d *DraftAvailablePlayer) GetId() common.RecordId {
	return d.ID
}

// SetId sets the unique identifier for this DraftAvailablePlayer record.
func (d *DraftAvailablePlayer) SetId(id common.RecordId) {
	d.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for DraftAvailablePlayer.
func (d *DraftAvailablePlayer) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that both the referenced Draft and User records exist.
func (d *DraftAvailablePlayer) DynamicallyValid(db common.DatabaseProvider) error {
	if err := common.ExistsById(db, &Draft{}, d.DraftId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(db, &User{}, d.PlayerId.RecordId()); err != nil {
		return err
	}
	return nil
}

// AccessibleTo returns everyone as DraftAvailablePlayer records are public.
func (d *DraftAvailablePlayer) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

// EditableBy returns only sysadmins as only they can modify DraftAvailablePlayer records.
func (d *DraftAvailablePlayer) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.SysAdminRecordId}
}

// Timestamps returns the create and update timestamps for this DraftAvailablePlayer record.
func (d *DraftAvailablePlayer) Timestamps() (time.Time, time.Time, *time.Time) {
	return d.CreatedAt, d.UpdatedAt, nil
}

// SetCreatedAt sets the creation timestamp for this DraftAvailablePlayer record.
func (d *DraftAvailablePlayer) SetCreatedAt(createdAt time.Time) {
	d.CreatedAt = createdAt
}

// SetUpdatedAt sets the update timestamp for this DraftAvailablePlayer record.
func (d *DraftAvailablePlayer) SetUpdatedAt(updatedAt time.Time) {
	d.UpdatedAt = updatedAt
}

func (d *DraftAvailablePlayer) BlankRecord() common.CrudRecord {
	return new(DraftAvailablePlayer)
}
