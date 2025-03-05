package model

import (
	"errors"
	"fmt"
	"intraclub/common"
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

type Format struct {
	ID     FormatId
	UserId UserId
	Lines  []Line
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

	for i, line1 := range f.Lines {
		for j, line2 := range f.Lines {
			if i != j && line1.EquivalentTo(line2) {
				return fmt.Errorf("format has duplicate lines %s at index %d, %s at index %d", line1, i, line2, j)
			}
		}
	}

	return nil
}

func (f *Format) DynamicallyValid(db common.DatabaseProvider) error {
	for _, line := range f.Lines {
		err := line.DynamicallyValid(db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Format) IsRatingValidForFormat(r RatingId) bool {
	for _, line := range f.Lines {
		if line.Player1Rating == r || line.Player2Rating == r {
			return true
		}
	}
	return false
}

func (f *Format) GetAvailableRatings() []RatingId {
	ratings := make([]RatingId, 0)
	for _, line := range f.Lines {
		if !ratingInList(line.Player1Rating, ratings) {
			ratings = append(ratings, line.Player1Rating)
		}
		if !ratingInList(line.Player2Rating, ratings) {
			ratings = append(ratings, line.Player2Rating)
		}
	}
	return ratings
}
