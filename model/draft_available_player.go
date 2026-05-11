package model

import (
	"context"
	"time"

	"intraclub/database"
)

// DraftAvailablePlayer is a join table record that links a Draft to available players.
// Each record represents a User who is eligible to be drafted in a specific Draft.
type DraftAvailablePlayer struct {
	ID        database.RecordId `json:"id"`
	DraftId   DraftId           `json:"draft_id"`
	PlayerId  UserId            `json:"player_id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as DraftAvailablePlayer has no specific owner.
func (d *DraftAvailablePlayer) GetOwner() database.RecordId {
	return database.InvalidRecordId
}

// SetOwner is a no-op as DraftAvailablePlayer has no specific owner.
func (d *DraftAvailablePlayer) SetOwner(recordId database.RecordId) {}

// Type returns the record type identifier for DraftAvailablePlayer.
func (d *DraftAvailablePlayer) Type() string {
	return "draft_available_player"
}

// GetId returns the unique identifier for this DraftAvailablePlayer record.
func (d *DraftAvailablePlayer) GetId() database.RecordId {
	return d.ID
}

// SetId sets the unique identifier for this DraftAvailablePlayer record.
func (d *DraftAvailablePlayer) SetId(id database.RecordId) {
	d.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for DraftAvailablePlayer.
func (d *DraftAvailablePlayer) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that both the referenced Draft and User records exist.
func (d *DraftAvailablePlayer) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	if err := database.ExistsById(ctx, db, &Draft{}, d.DraftId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &User{}, d.PlayerId.RecordId()); err != nil {
		return err
	}
	return nil
}

func (d *DraftAvailablePlayer) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return database.AccessibleToEveryone
}

func (d *DraftAvailablePlayer) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return []database.RecordId{database.SysAdminRecordId}
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

func (d *DraftAvailablePlayer) BlankRecord() database.CrudRecord {
	return new(DraftAvailablePlayer)
}
