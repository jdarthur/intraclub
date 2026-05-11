package model

import (
	"context"
	"time"

	"intraclub/database"
)

// DraftPick is a join table record that represents a single pick in a Draft.
// Each record tracks which user was selected by which team, in which round and pick number,
// and includes the player's rating at the time of the draft.
type DraftPick struct {
	ID        database.RecordId `json:"id"`
	DraftId   DraftId           `json:"draft_id"`
	TeamId    TeamId            `json:"team_id"`
	UserId    UserId            `json:"user_id"`
	Round     int               `json:"round"`
	Pick      int               `json:"pick"`
	Rating    RatingId          `json:"rating"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as DraftPick has no specific owner.
func (d *DraftPick) GetOwner() database.RecordId {
	return database.InvalidRecordId
}

// SetOwner is a no-op as DraftPick has no specific owner.
func (d *DraftPick) SetOwner(recordId database.RecordId) {}

// Type returns the record type identifier for DraftPick.
func (d *DraftPick) Type() string {
	return "draft_pick"
}

// GetId returns the unique identifier for this DraftPick record.
func (d *DraftPick) GetId() database.RecordId {
	return d.ID
}

// SetId sets the unique identifier for this DraftPick record.
func (d *DraftPick) SetId(id database.RecordId) {
	d.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for DraftPick.
func (d *DraftPick) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that the referenced Draft, Team, and User records all exist.
func (d *DraftPick) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	if err := database.ExistsById(ctx, db, &Draft{}, d.DraftId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &Team{}, d.TeamId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &User{}, d.UserId.RecordId()); err != nil {
		return err
	}
	return nil
}

func (d *DraftPick) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return database.AccessibleToEveryone
}

func (d *DraftPick) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return []database.RecordId{database.SysAdminRecordId}
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

func (d *DraftPick) BlankRecord() database.CrudRecord {
	return new(DraftPick)
}
