package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

// DraftFormat is a join table record that links a Draft to its Format.
// This enables efficient reverse lookups from Format to Drafts without
// scanning the entire Draft collection.
type DraftFormat struct {
	ID       database.RecordId `json:"id"`
	DraftId  DraftId           `json:"draft_id"`
	FormatId FormatId          `json:"format_id"`
}

func (d *DraftFormat) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (d *DraftFormat) SetOwner(userId database.UserId) {}

func (d *DraftFormat) Type() string {
	return "draft_format"
}

func (d *DraftFormat) GetId() database.RecordId {
	return d.ID
}

func (d *DraftFormat) SetId(id database.RecordId) {
	d.ID = id
}

func (d *DraftFormat) StaticallyValid() error {
	return nil
}

func (d *DraftFormat) DynamicallyValid(ctx context.Context, db database.Provider) error {
	_, err := database.GetExistingRecordById(ctx, db, &Draft{}, d.DraftId.RecordId())
	if err != nil {
		return err
	}

	_, err = database.GetExistingRecordById(ctx, db, &Format{}, d.FormatId.RecordId())
	if err != nil {
		return err
	}

	return nil
}

func (d *DraftFormat) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (d *DraftFormat) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (d *DraftFormat) NewRecord() database.CrudRecord {
	return new(DraftFormat)
}

func (d *DraftFormat) UniquenessEquivalent(other *DraftFormat) error {
	if d.DraftId == other.DraftId {
		return fmt.Errorf("draft %s already has an assigned format", d.DraftId)
	}
	return nil
}
