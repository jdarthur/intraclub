package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newStoredFormat(t *testing.T, db common.DatabaseProvider, lines []FormatLine) *Format {
	f := NewFormat()
	f.Lines = lines

	v, err := common.CreateOne(db, f)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newLine(t *testing.T, db common.DatabaseProvider) FormatLine {
	return FormatLine{
		Player1Rating: newStoredRating(t, db).ID,
		Player2Rating: newStoredRating(t, db).ID,
	}
}

func newDefaultFormat(t *testing.T, db common.DatabaseProvider) *Format {
	user := newStoredUser(t, db)
	f := NewFormat()
	f.UserId = user.ID
	lines := []FormatLine{
		newLine(t, db),
		newLine(t, db),
	}

	f.Name = "default format"
	f.Lines = lines
	f.PossibleRatings = []RatingId{
		lines[0].Player1Rating,
		lines[0].Player2Rating,
		lines[1].Player1Rating,
		lines[1].Player2Rating,
	}
	return f
}

func newDefaultStoredFormat(t *testing.T, db common.DatabaseProvider) *Format {
	f := newDefaultFormat(t, db)
	v, err := common.CreateOne(db, f)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestFormatDuplicateLine(t *testing.T) {
	db := common.NewUnitTestDBProvider()

	format := newDefaultFormat(t, db)
	format.Lines = []FormatLine{format.Lines[0], format.Lines[0]}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected duplicate line to fail")
	}
	fmt.Println(err)
}

func TestFormatReversedValueDuplicateLine(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultFormat(t, db)

	line1 := format.Lines[0]
	line2 := FormatLine{Player1Rating: line1.Player2Rating, Player2Rating: line1.Player1Rating}

	format.Lines = []FormatLine{line1, line2}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected duplicate line to fail")
	}
	fmt.Println(err)
}

func TestFormatNameEmpty(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.Name = ""
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected empty name to fail")
	}
	fmt.Println(err)
}

func TestFormatNameWhitespace(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.Name = "   "
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected whitespace name to fail")
	}
	fmt.Println(err)
}

func TestFormatHasEmptyPossibleRatings(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.PossibleRatings = []RatingId{}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected empty possible ratings to fail")
	}
	fmt.Println(err)
}

func TestFormatHasInvalidUserId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultFormat(t, db)

	format.UserId = UserId(common.InvalidRecordId)
	err := format.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected invalid user id to fail")
	}
	fmt.Println(err)
}

func TestFormatHasEmptyLines(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.Lines = []FormatLine{}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected empty lines to fail")
	}
	fmt.Println(err)
}

func TestFormatHasLineRatingsNotInPossibleLinesList(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.Lines = []FormatLine{
		newLine(t, db),
	}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected empty lines to fail")
	}
	fmt.Println(err)
}

func TestFormatCannotBeDeletedWhenInUse(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newDefaultStoredDraft(t, db)

	_, _, err := common.DeleteOneById(db, &Format{}, draft.Format.RecordId())
	if err == nil {
		t.Fatal("Expected in-use format delete to fail")
	}
	fmt.Println(err)
}

func TestFormatCannotBeEditedWhenInUse(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newDefaultStoredDraft(t, db)

	format, err := common.GetExistingRecordById(db, &Format{}, draft.Format.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	newRating := newStoredRating(t, db)
	format.PossibleRatings = append(format.PossibleRatings, newRating.ID)
	err = common.UpdateOne(db, format)
	if err == nil {
		t.Fatal("Expected in-use format edit to fail")
	}
	fmt.Println(err)
}

func TestFormatCanBeEditedWhenNotInUse(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	f := newDefaultStoredFormat(t, db)

	newRating := newStoredRating(t, db)
	f.PossibleRatings = append(f.PossibleRatings, newRating.ID)

	err := common.UpdateOne(db, f)
	if err != nil {
		t.Fatal(err)
	}
}
