package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

// FormatRating is a join table record that links a Format to a Rating.
// This enables efficient reverse lookups from Rating to Formats without
// scanning the entire Format collection.
type FormatRating struct {
	ID       database.RecordId `json:"id"`
	FormatId FormatId          `json:"format_id"`
	RatingId RatingId          `json:"rating_id"`
}

func (f *FormatRating) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (f *FormatRating) SetOwner(userId database.UserId) {}

func (f *FormatRating) Type() string {
	return "format_rating"
}

func (f *FormatRating) GetId() database.RecordId {
	return f.ID
}

func (f *FormatRating) SetId(id database.RecordId) {
	f.ID = id
}

func (f *FormatRating) StaticallyValid() error {
	return nil
}

func (f *FormatRating) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	_, err := database.GetExistingRecordById(ctx, db, &Format{}, f.FormatId.RecordId())
	if err != nil {
		return err
	}

	_, err = database.GetExistingRecordById(ctx, db, &Rating{}, f.RatingId.RecordId())
	if err != nil {
		return err
	}

	return nil
}

func (f *FormatRating) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return database.AccessibleToEveryone
}

func (f *FormatRating) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (f *FormatRating) BlankRecord() database.CrudRecord {
	return new(FormatRating)
}

func (f *FormatRating) UniquenessEquivalent(other *FormatRating) error {
	if f.FormatId == other.FormatId && f.RatingId == other.RatingId {
		return fmt.Errorf("format %s already has rating %s assigned", f.FormatId, f.RatingId)
	}
	return nil
}
