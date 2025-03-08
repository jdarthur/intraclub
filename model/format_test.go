package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newStoredFormat(t *testing.T, db common.DatabaseProvider, lines []Line) *Format {
	f := NewFormat()
	f.Lines = lines

	v, err := common.CreateOne(db, f)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newLine(t *testing.T, db common.DatabaseProvider) Line {
	return Line{
		Player1Rating: newStoredRating(t, db).ID,
		Player2Rating: newStoredRating(t, db).ID,
	}
}

func newDefaultStoredFormat(t *testing.T, db common.DatabaseProvider) *Format {
	f := NewFormat()
	lines := []Line{newLine(t, db)}

	f.Name = "default format"
	f.Lines = lines
	f.PossibleRatings = []RatingId{lines[0].Player1Rating, lines[0].Player2Rating}

	v, err := common.CreateOne(db, f)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestFormatDuplicateLine(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	line1 := newLine(t, db)

	format := NewFormat()
	format.Lines = []Line{line1, line1}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected duplicate line to fail")
	}
	fmt.Println(err)
}

func TestFormatReversedValueDuplicateLine(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	line1 := newLine(t, db)
	line2 := Line{Player1Rating: line1.Player2Rating, Player2Rating: line1.Player1Rating}

	format := NewFormat()
	format.Lines = []Line{line1, line2}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected duplicate line to fail")
	}
	fmt.Println(err)
}

func TestFormatNameEmpty(t *testing.T) {
	t.Fatal("implement me")
}

func TestFormatNameWhitespace(t *testing.T) {
	t.Fatal("implement me")
}

func TestFormatHasEmptyPossibleRatings(t *testing.T) {
	t.Fatal("implement me")
}

func TestFormatHasEmptyLines(t *testing.T) {
	t.Fatal("implement me")
}

func TestFormatHasLineRatingsNotInPossibleLinesList(t *testing.T) {
	t.Fatal("implement me")
}

func TestFormatCannotBeDeletedWhenInUse(t *testing.T) {
	t.Fatal("implement me")
}

func TestFormatCannotBeEditedWhenInUse(t *testing.T) {
	t.Fatal("implement me")
}

func TestFormatCanBeEditedWhenNotInUse(t *testing.T) {
	t.Fatal("implement me")
}
