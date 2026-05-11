package model

import (
	"context"
	"fmt"
	"time"

	"intraclub/database"
)

// DraftCaptain is a join table record that links a Draft to its team captains.
// Each record represents a Team-Captain assignment for a specific Draft.
type DraftCaptain struct {
	ID         database.RecordId `json:"id"`
	DraftId    DraftId           `json:"draft_id"`
	TeamId     TeamId            `json:"team_id"`
	CaptainId  UserId            `json:"captain_id"`
	DraftOrder int               `json:"draft_order"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// GetOwner returns InvalidRecordId as DraftCaptain has no specific owner.
func (d *DraftCaptain) GetOwner() database.RecordId {
	return database.InvalidRecordId
}

// SetOwner is a no-op as DraftCaptain has no specific owner.
func (d *DraftCaptain) SetOwner(recordId database.RecordId) {}

// Type returns the record type identifier for DraftCaptain.
func (d *DraftCaptain) Type() string {
	return "draft_captain"
}

// GetId returns the unique identifier for this DraftCaptain record.
func (d *DraftCaptain) GetId() database.RecordId {
	return d.ID
}

// SetId sets the unique identifier for this DraftCaptain record.
func (d *DraftCaptain) SetId(id database.RecordId) {
	d.ID = id
}

// StaticallyValid always returns nil as there are no static validation rules for DraftCaptain.
func (d *DraftCaptain) StaticallyValid() error {
	return nil
}

// DynamicallyValid verifies that the referenced Draft, Team, and User records all exist.
func (d *DraftCaptain) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {

	// ensure tha the user ID exists
	err := database.ExistsById(ctx, db, &User{}, d.CaptainId.RecordId())
	if err != nil {
		return err
	}

	// ensure that the draft exists
	_, err = database.GetExistingRecordById(ctx, db, &Draft{}, d.DraftId.RecordId())
	if err != nil {
		return err
	}

	// ensure that the team exists
	team, err := database.GetExistingRecordById(ctx, db, &Team{}, d.TeamId.RecordId())
	if err != nil {
		return err
	}

	// ensure that this user is the actual captain of the referenced team
	captain, err := team.GetCaptain(ctx, db)
	if err != nil {
		return err
	} else if captain != d.CaptainId {
		return fmt.Errorf("user %s is not the captain of team %s (expected %s)", d.CaptainId, d.TeamId.RecordId(), captain)
	}
	return nil
}

func (d *DraftCaptain) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return database.AccessibleToEveryone
}

func (d *DraftCaptain) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return []database.RecordId{database.SysAdminRecordId}
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

func (d *DraftCaptain) BlankRecord() database.CrudRecord {
	return new(DraftCaptain)
}
