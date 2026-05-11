package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"intraclub/database"
)

// FormatLine is a pairing of two players that have a particular Rating.
// Each week, a given Format will be composed of one or more FormatLine s
// from which each Team will compose a Lineup. A Lineup from one
// team on a given Week in a Season will play a Lineup from the
// opposing Team based on the Schedule
type FormatLine struct {
	Player1Rating RatingId `json:"player_1_rating"` // ID of the Rating record for player two in this FormatLine
	Player2Rating RatingId `json:"player_2_rating"` // ID of the Rating record for player one in this FormatLine
}

func (l *FormatLine) UnmarshalJSON(bytes []byte) error {
	m := map[string]string{}
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}
	rating1, err := database.RecordIdFromString(m["player_1_rating"])
	if err != nil {
		return err
	}
	rating2, err := database.RecordIdFromString(m["player_2_rating"])
	if err != nil {
		return err
	}

	l.Player1Rating = RatingId(rating1)
	l.Player2Rating = RatingId(rating2)
	return nil
}

func (l *FormatLine) EquivalentTo(other FormatLine) bool {
	if l.Player1Rating == other.Player1Rating && l.Player2Rating == other.Player2Rating {
		return true
	}

	if l.Player1Rating == other.Player2Rating && l.Player2Rating == other.Player1Rating {
		return true
	}

	return false
}

func (l *FormatLine) StaticallyValid() error {
	return nil
}

func (l *FormatLine) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	err := database.ExistsById(ctx, db, &Rating{}, l.Player1Rating.RecordId())
	if err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &Rating{}, l.Player2Rating.RecordId())
}

func (l *FormatLine) String() string {
	return fmt.Sprintf("%s / %s", l.Player1Rating, l.Player2Rating)
}

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
// a user to specify a format that a Season will be played in. This is
// composed of a list of possible Rating IDs, and a list of FormatLine records
// which compose a pairing of two Rating types.
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
	ID              FormatId     `json:"id"`               // unique ID for the Format
	UserId          database.UserId       `json:"user_id"`          // owner of the Format
	Name            string       `json:"name"`             // name for the Format, e.g. "Men's Intraclub 1/2/3"
	PossibleRatings RatingList   `json:"possible_ratings"` // list of possible Rating values for the lines, highest to lowest skill
	Lines           []FormatLine `json:"lines"`            // Rating pairings that will play during a matchup
}

func (f *Format) GetOwner() database.UserId {
	return f.UserId
}

func (f *Format) PreUpdate(ctx context.Context, db database.DatabaseProvider, existingValues database.CrudRecord) error {
	return f.CheckHasAssignedDrafts(ctx, db, true)
}

func (f *Format) PreDelete(ctx context.Context, db database.DatabaseProvider) error {
	return f.CheckHasAssignedDrafts(ctx, db, false)
}

func (f *Format) SetOwner(userId database.UserId) {
	f.UserId = userId
}

func NewFormat() *Format {
	return &Format{}
}

func (f *Format) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.UserId {
	return []database.UserId{f.UserId}
}

func (f *Format) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.UserId {
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

func (f *Format) StaticallyValid() error {
	if len(f.Lines) == 0 {
		return errors.New("format has no lines")
	}

	if len(f.PossibleRatings) == 0 {
		return errors.New("format has no possible ratings")
	}

	f.Name = strings.TrimSpace(f.Name)
	if f.Name == "" {
		return errors.New("format has no name")
	}

	fmt.Println(f)

	for i, line1 := range f.Lines {
		if !f.IsRatingInOptionsList(line1.Player1Rating) {
			return fmt.Errorf("rating for player 1 in line %d (%s) is not in possible options list", i, line1.Player1Rating)
		}
		if !f.IsRatingInOptionsList(line1.Player2Rating) {
			return fmt.Errorf("rating for player 2 in line %d (%s) is not in possible options list", i, line1.Player1Rating)
		}
		for j, line2 := range f.Lines {
			if i != j && line1.EquivalentTo(line2) {
				return fmt.Errorf("format has duplicate lines %s at index %d, %s at index %d", line1, i, line2, j)
			}
		}
	}

	return nil
}

func (f *Format) IsRatingInOptionsList(r RatingId) bool {
	for _, option := range f.PossibleRatings {
		if r == option {
			return true
		}
	}
	return false
}

func (f *Format) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	err := database.ExistsById(ctx, db, &User{}, f.UserId.RecordId())
	if err != nil {
		return err
	}

	for _, line := range f.Lines {
		err := line.DynamicallyValid(ctx, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Format) IsRatingValidForFormat(r RatingId) bool {
	for _, rating := range f.PossibleRatings {
		if r == rating {
			return true
		}
	}
	return false
}

func (f *Format) GetAssignedDrafts(ctx context.Context, db database.DatabaseProvider) ([]*Draft, error) {
	return database.GetAllWhere[*Draft](ctx, db, func(_ context.Context, c *Draft) bool {
		return c.Format == f.ID
	})
}

func (f *Format) CheckHasAssignedDrafts(ctx context.Context, db database.DatabaseProvider, isUpdate bool) error {
	drafts, err := f.GetAssignedDrafts(ctx, db)
	if err != nil {
		return err
	}

	verb := "edit"
	if !isUpdate {
		verb = "delete"
	}

	if len(drafts) != 0 {
		return fmt.Errorf("cannot %s format with %d assigned drafts", verb, len(drafts))
	}
	return nil
}

func (f *Format) BlankRecord() database.CrudRecord {
	return new(Format)
}
