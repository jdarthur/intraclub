package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"strings"
)

// Line is a pairing of two players that have a particular Rating.
// Each week, a given Format will be composed of one or more Line s
// from which each Team will compose a Lineup. A Lineup from one
// team on a given Week in a Season will play a Lineup from the
// opposing Team based on the Schedule
type Line struct {
	Player1Rating RatingId // ID of the Rating record for player two in this Line
	Player2Rating RatingId // ID of the Rating record for player one in this Line
}

func (l Line) EquivalentTo(other Line) bool {
	if l.Player1Rating == other.Player1Rating && l.Player2Rating == other.Player2Rating {
		return true
	}

	if l.Player1Rating == other.Player2Rating && l.Player2Rating == other.Player1Rating {
		return true
	}

	return false
}

func (l Line) StaticallyValid() error {
	return nil
}

func (l Line) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &Rating{}, l.Player1Rating.RecordId())
	if err != nil {
		return err
	}
	return common.ExistsById(db, &Rating{}, l.Player2Rating.RecordId())
}

func (l Line) String() string {
	return fmt.Sprintf("%s / %s", l.Player1Rating, l.Player2Rating)
}

type FormatId common.RecordId

func (id FormatId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id FormatId) String() string {
	return id.RecordId().String()
}

// Format is a globally-available common.CrudRecord type which allows
// a user to specify a format that a Season will be played in. This is
// composed of a list of possible Rating IDs, and a list of Line records
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
// Line options of:
//   - old guy / old guy
//   - old guy / young guy
//   - young guy / young guy.
type Format struct {
	ID              FormatId   // unique ID for the Format
	UserId          UserId     // owner of the Format
	Name            string     // name for the Format, e.g. "Men's Intraclub 1/2/3"
	PossibleRatings []RatingId // list of possible Rating values for the lines, highest to lowest skill
	Lines           []Line     // Rating pairings that will play during a matchup
}

func (f *Format) PreUpdate(db common.DatabaseProvider, existingValues common.CrudRecord) error {
	return f.CheckHasAssignedDrafts(db, true)
}

func (f *Format) PreDelete(db common.DatabaseProvider) error {
	return f.CheckHasAssignedDrafts(db, false)
}

func (f *Format) SetOwner(recordId common.RecordId) {
	f.UserId = UserId(recordId)
}

func NewFormat() *Format {
	return &Format{}
}

func (f *Format) EditableBy(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{f.UserId.RecordId()}
}

func (f *Format) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (f *Format) Type() string {
	return "format"
}

func (f *Format) GetId() common.RecordId {
	return f.ID.RecordId()
}

func (f *Format) SetId(id common.RecordId) {
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

func (f *Format) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &User{}, f.UserId.RecordId())
	if err != nil {
		return err
	}

	for _, line := range f.Lines {
		err := line.DynamicallyValid(db)
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

func (f *Format) GetAssignedDrafts(db common.DatabaseProvider) ([]*Draft, error) {
	return common.GetAllWhere(db, &Draft{}, func(c *Draft) bool {
		return c.Format == f.ID
	})
}

func (f *Format) CheckHasAssignedDrafts(db common.DatabaseProvider, isUpdate bool) error {
	drafts, err := f.GetAssignedDrafts(db)
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
