package model

import (
	"context"
	"fmt"

	"intraclub/database"
)

// DraftRatingCutoff is a join table record that assigns a rating to a cutoff
// index (the last selection index matching that rating) for a particular Draft.
// This replaces the former denormalized `Draft.RatingCutoffs` inline map
// (rating -> cutoff index), enabling the relationship to be queried/indexed
// individually and stored in its own collection/table.
//
// It is stored in its own collection (`draft_rating_cutoff`) with FKs to draft
// and rating, and a natural unique constraint on (DraftId, RatingId). When the
// #36 SQLite provider lands, this becomes a `draft_rating_cutoffs` table with a
// migration/backfill.
type DraftRatingCutoff struct {
	ID          database.RecordId `json:"id"`
	DraftId     DraftId           `json:"draft_id"`
	RatingId    RatingId          `json:"rating_id"`
	CutoffIndex int               `json:"cutoff_index"`
}

func (d *DraftRatingCutoff) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (d *DraftRatingCutoff) SetOwner(userId database.UserId) {}

func (d *DraftRatingCutoff) Type() string {
	return "draft_rating_cutoff"
}

func (d *DraftRatingCutoff) GetId() database.RecordId {
	return d.ID
}

func (d *DraftRatingCutoff) SetId(id database.RecordId) {
	d.ID = id
}

func (d *DraftRatingCutoff) StaticallyValid() error {
	return nil
}

func (d *DraftRatingCutoff) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Draft{}, d.DraftId.RecordId()); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &Rating{}, d.RatingId.RecordId())
}

func (d *DraftRatingCutoff) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (d *DraftRatingCutoff) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (d *DraftRatingCutoff) NewRecord() database.CrudRecord {
	return new(DraftRatingCutoff)
}

// UniquenessEquivalent enforces the natural unique constraint on
// (DraftId, RatingId): a rating may only be assigned one cutoff per draft.
func (d *DraftRatingCutoff) UniquenessEquivalent(other *DraftRatingCutoff) error {
	if d.DraftId == other.DraftId && d.RatingId == other.RatingId {
		return fmt.Errorf("draft %s already has a rating cutoff assigned for rating %s", d.DraftId, d.RatingId)
	}
	return nil
}
