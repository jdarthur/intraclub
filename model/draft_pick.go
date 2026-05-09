package model

import (
	"intraclub/common"
	"time"
)

// DraftPick is a join table record that represents a single pick in a Draft.
// Each record tracks which user was selected by which team, in which round and pick number,
// and includes the player's rating at the time of the draft.
type DraftPick struct {
	ID        common.RecordId `json:"id"`
	DraftId   DraftId         `json:"draft_id"`
	TeamId    TeamId          `json:"team_id"`
	UserId    UserId          `json:"user_id"`
	Round     int             `json:"round"`
	Pick      int             `json:"pick"`
	Rating    RatingId        `json:"rating"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as DraftPick has no specific owner.
func (d *DraftPick) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

// SetOwner is a no-op as DraftPick has no specific owner.
func (d *DraftPick) SetOwner(recordId common.RecordId) {}

// Type returns the record type identifier for DraftPick.
func (d *DraftPick) Type() string {
	return "draft_pick"
}

// GetId returns the unique identifier for this DraftPick record.
func (d *DraftPick) GetId() common.RecordId {
	return d.ID
}

// SetId sets the unique identifier for this DraftPick record.
func (d *DraftPick) SetId(id common.RecordId) {
	d.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for DraftPick.
func (d *DraftPick) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that the referenced Draft, Team, and User records all exist.
func (d *DraftPick) DynamicallyValid(db common.DatabaseProvider) error {
	if err := common.ExistsById(db, &Draft{}, d.DraftId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(db, &Team{}, d.TeamId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(db, &User{}, d.UserId.RecordId()); err != nil {
		return err
	}
	return nil
}

// AccessibleTo returns everyone as DraftPick records are public.
func (d *DraftPick) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

// EditableBy returns only sysadmins as only they can modify DraftPick records.
func (d *DraftPick) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.SysAdminRecordId}
}

// Timestamps returns the create and update timestamps for this DraftPick record.
func (d *DraftPick) Timestamps() (time.Time, time.Time, *time.Time) {
	return d.CreatedAt, d.UpdatedAt, nil
}

// SetCreatedAt sets the creation timestamp for this DraftPick record.
func (d *DraftPick) SetCreatedAt(createdAt time.Time) {
	d.CreatedAt = createdAt
}

// SetUpdatedAt sets the update timestamp for this DraftPick record.
func (d *DraftPick) SetUpdatedAt(updatedAt time.Time) {
	d.UpdatedAt = updatedAt
}

func (d *DraftPick) BlankRecord() common.CrudRecord {
	return new(DraftPick)
}
