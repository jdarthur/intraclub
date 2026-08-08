package model

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"intraclub/database"
)

type FormatId database.RecordId

func (id FormatId) UnmarshalJSON(bytes []byte) error {
	rid := id.RecordId()
	return (*database.RecordId)(&rid).UnmarshalJSON(bytes)
}

func (id FormatId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id FormatId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id FormatId) String() string {
	return id.RecordId().String()
}

// Format is a globally-available common.CrudRecord type which allows
// a user to specify a format that a Season will be played in. It is composed
// of a list of possible Rating IDs, and a list of FormatLine records which
// compose a pairing of two Rating types.
//
// For example, this could be a 1/2/3 division of skilled, medium, and
// beginner-level players with all six combinations of skill level:
//   - 1/1
//   - 1/2
//   - 1/3
//   - 2/2
//   - 2/3
//   - 3/3
//
// Another format type could be "old guy / young guy" in which players
// are classed into either "old guy" status or "young guy" status, with
// FormatLine options of:
//   - old guy / old guy
//   - old guy / young guy
//   - young guy / young guy.
type Format struct {
	ID     FormatId        `json:"id"`      // unique ID for the Format
	UserId database.UserId `json:"user_id"` // owner of the Format
	Name   string          `json:"name"`    // name for the Format, e.g. "Men's Intraclub 1/2/3"
}

// API/JSON shape decision: the former inline `possible_ratings`
// (`RatingList`) and `lines` (`[]FormatLine`) fields have been removed from
// `Format` and normalized into the `FormatRating` and `FormatLine` join tables
// (format_rating / format_line collections). In-process reads reassemble the
// relationships in order via `Format.GetPossibleRatings` and `Format.GetLines`.
// This matches the established join-table normalization (see #58 for the same
// treatment of Draft.RatingCutoffs) and the schema conventions in
// docs/schema-conventions.md.

func (f *Format) GetOwner() database.UserId {
	return f.UserId
}

func (f *Format) PreUpdate(ctx context.Context, db database.Provider, existingValues database.CrudRecord) error {
	return f.CheckHasAssignedDrafts(ctx, db, true)
}

func (f *Format) PreDelete(ctx context.Context, db database.Provider) error {
	return f.CheckHasAssignedDrafts(ctx, db, false)
}

func (f *Format) SetOwner(userId database.UserId) {
	f.UserId = userId
}

func NewFormat() *Format {
	return &Format{}
}

func (f *Format) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{f.UserId}
}

func (f *Format) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (f *Format) Type() string {
	return "format"
}

func (f *Format) GetId() database.RecordId {
	return f.ID.RecordId()
}

func (f *Format) SetId(id database.RecordId) {
	f.ID = FormatId(id)
}

// StaticallyValid validates the Format's scalar fields. The possible-ratings
// and lines content (now stored in child tables) is validated by
// DynamicallyValid, which can query those relationships.
func (f *Format) StaticallyValid() error {
	f.Name = strings.TrimSpace(f.Name)
	if f.Name == "" {
		return errors.New("format has no name")
	}
	return nil
}

// GetPossibleRatings returns this format's possible ratings, ordered
// highest-skill to lowest-skill (i.e. by FormatRating.RatingIndex).
func (f *Format) GetPossibleRatings(ctx context.Context, db database.Provider) (RatingList, error) {
	formatRatings, err := database.GetAllWhere[*FormatRating](ctx, db, func(_ context.Context, fr *FormatRating) bool {
		return fr.FormatId == f.ID
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(formatRatings, func(a, b *FormatRating) int {
		return a.RatingIndex - b.RatingIndex
	})
	ratings := make(RatingList, 0, len(formatRatings))
	for _, fr := range formatRatings {
		ratings = append(ratings, fr.RatingId)
	}
	return ratings, nil
}

// SetPossibleRatings replaces this format's possible ratings with the provided
// list, preserving order as RatingIndex values in the FormatRating join table.
// The format must already have an assigned ID (i.e. be created). An empty list
// is rejected because a format requires at least one possible rating.
func (f *Format) SetPossibleRatings(ctx context.Context, db database.Provider, ratings RatingList) error {
	if len(ratings) == 0 {
		return errors.New("format has no possible ratings")
	}
	existing, err := database.GetAllWhere[*FormatRating](ctx, db, func(_ context.Context, fr *FormatRating) bool {
		return fr.FormatId == f.ID
	})
	if err != nil {
		return err
	}
	for _, fr := range existing {
		if _, _, err := database.DeleteOneById(ctx, db, fr, fr.ID); err != nil {
			return err
		}
	}
	for i, ratingId := range ratings {
		if _, err := database.CreateOne(ctx, db, &FormatRating{
			FormatId:    f.ID,
			RatingId:    ratingId,
			RatingIndex: i,
		}); err != nil {
			return err
		}
	}
	return nil
}

// GetLines returns this format's lines, ordered by FormatLine.FormatIndex.
func (f *Format) GetLines(ctx context.Context, db database.Provider) ([]FormatLine, error) {
	lines, err := database.GetAllWhere[*FormatLine](ctx, db, func(_ context.Context, l *FormatLine) bool {
		return l.FormatId == f.ID
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(lines, func(a, b *FormatLine) int {
		return a.FormatIndex - b.FormatIndex
	})
	result := make([]FormatLine, 0, len(lines))
	for _, l := range lines {
		result = append(result, *l)
	}
	return result, nil
}

// SetLines replaces this format's lines with the provided list, preserving
// order as FormatIndex values in the FormatLine join table. The format must
// already have an assigned ID (i.e. be created). It rejects an empty list,
// duplicate / reversed-duplicate lines, and lines whose ratings are not among
// the format's possible ratings.
func (f *Format) SetLines(ctx context.Context, db database.Provider, lines []FormatLine) error {
	if len(lines) == 0 {
		return errors.New("format has no lines")
	}

	possibleRatings, err := f.GetPossibleRatings(ctx, db)
	if err != nil {
		return err
	}
	for i, line1 := range lines {
		if !IsRatingInOptionsList(possibleRatings, line1.Player1Rating) {
			return fmt.Errorf("rating for player 1 in line %d (%s) is not in possible options list", i, line1.Player1Rating)
		}
		if !IsRatingInOptionsList(possibleRatings, line1.Player2Rating) {
			return fmt.Errorf("rating for player 2 in line %d (%s) is not in possible options list", i, line1.Player2Rating)
		}
		for j, line2 := range lines {
			if i != j && line1.EquivalentTo(line2) {
				return fmt.Errorf("format has duplicate lines %s at index %d, %s at index %d", line1.String(), i, line2.String(), j)
			}
		}
	}

	existing, err := database.GetAllWhere[*FormatLine](ctx, db, func(_ context.Context, l *FormatLine) bool {
		return l.FormatId == f.ID
	})
	if err != nil {
		return err
	}
	for _, l := range existing {
		if _, _, err := database.DeleteOneById(ctx, db, l, l.ID); err != nil {
			return err
		}
	}
	for i, line := range lines {
		if _, err := database.CreateOne(ctx, db, &FormatLine{
			FormatId:      f.ID,
			FormatIndex:   i,
			Player1Rating: line.Player1Rating,
			Player2Rating: line.Player2Rating,
		}); err != nil {
			return err
		}
	}
	return nil
}

// DynamicallyValid validates the format against its relationships: the owner
// exists and, when the format has content (possible ratings / lines set via the
// Set* methods), that content is well-formed. A freshly-created format starts
// with empty child tables and is completed by SetPossibleRatings/SetLines, so
// the non-empty/duplicate/options checks live in those setters rather than
// here (mirroring how Draft validates rating cutoffs once they are present).
func (f *Format) DynamicallyValid(ctx context.Context, db database.Provider) error {
	err := database.ExistsById(ctx, db, &User{}, f.UserId.RecordId())
	if err != nil {
		return err
	}

	possibleRatings, err := f.GetPossibleRatings(ctx, db)
	if err != nil {
		return err
	}

	lines, err := f.GetLines(ctx, db)
	if err != nil {
		return err
	}

	if len(possibleRatings) > 0 && len(lines) > 0 {
		for i, line1 := range lines {
			if !IsRatingInOptionsList(possibleRatings, line1.Player1Rating) {
				return fmt.Errorf("rating for player 1 in line %d (%s) is not in possible options list", i, line1.Player1Rating)
			}
			if !IsRatingInOptionsList(possibleRatings, line1.Player2Rating) {
				return fmt.Errorf("rating for player 2 in line %d (%s) is not in possible options list", i, line1.Player2Rating)
			}
			for j, line2 := range lines {
				if i != j && line1.EquivalentTo(line2) {
					return fmt.Errorf("format has duplicate lines %s at index %d, %s at index %d", line1.String(), i, line2.String(), j)
				}
			}
		}
	}

	for _, line := range lines {
		if err := line.DynamicallyValid(ctx, db); err != nil {
			return err
		}
	}
	return nil
}

// IsRatingInOptionsList reports whether r appears in the given possible-ratings list.
func IsRatingInOptionsList(ratings RatingList, r RatingId) bool {
	for _, option := range ratings {
		if r == option {
			return true
		}
	}
	return false
}

// IsRatingValidForFormat reports whether r is one of this format's possible ratings.
func (f *Format) IsRatingValidForFormat(ctx context.Context, db database.Provider, r RatingId) (bool, error) {
	possibleRatings, err := f.GetPossibleRatings(ctx, db)
	if err != nil {
		return false, err
	}
	for _, rating := range possibleRatings {
		if r == rating {
			return true, nil
		}
	}
	return false, nil
}

func (f *Format) GetAssignedDrafts(ctx context.Context, db database.Provider) ([]*Draft, error) {
	draftFormats, err := database.GetAllWhere[*DraftFormat](ctx, db, func(_ context.Context, df *DraftFormat) bool {
		return df.FormatId == f.ID
	})
	if err != nil {
		return nil, err
	}

	drafts := make([]*Draft, 0, len(draftFormats))
	for _, df := range draftFormats {
		d, err := database.GetExistingRecordById(ctx, db, &Draft{}, df.DraftId.RecordId())
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, d)
	}
	return drafts, nil
}

func (f *Format) CheckHasAssignedDrafts(ctx context.Context, db database.Provider, isUpdate bool) error {
	draftFormats, err := database.GetAllWhere[*DraftFormat](ctx, db, func(_ context.Context, df *DraftFormat) bool {
		return df.FormatId == f.ID
	})
	if err != nil {
		return err
	}

	verb := "edit"
	if !isUpdate {
		verb = "delete"
	}

	if len(draftFormats) != 0 {
		return fmt.Errorf("cannot %s format with %d assigned drafts", verb, len(draftFormats))
	}
	return nil
}

func (f *Format) NewRecord() database.CrudRecord {
	return new(Format)
}
