package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredFormat(t *testing.T, db database.Provider, lines []FormatLine) *Format {
	f := NewFormat()
	f.Lines = lines

	v, err := database.CreateOne(context.Background(), db, f)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newLine(t *testing.T, db database.Provider) FormatLine {
	return FormatLine{
		Player1Rating: newStoredRating(t, db).ID,
		Player2Rating: newStoredRating(t, db).ID,
	}
}

func newDefaultFormat(t *testing.T, db database.Provider) *Format {
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

func newDefaultStoredFormat(t *testing.T, db database.Provider) *Format {
	f := newDefaultFormat(t, db)
	v, err := database.CreateOne(context.Background(), db, f)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestFormatDuplicateLine(t *testing.T) {
	db := database.NewUnitTestDBProvider()

	format := newDefaultFormat(t, db)
	format.Lines = []FormatLine{format.Lines[0], format.Lines[0]}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected duplicate line to fail")
	}
	fmt.Println(err)
}

func TestFormatReversedValueDuplicateLine(t *testing.T) {
	db := database.NewUnitTestDBProvider()
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
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.Name = ""
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected empty name to fail")
	}
	fmt.Println(err)
}

func TestFormatNameWhitespace(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.Name = "   "
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected whitespace name to fail")
	}
	fmt.Println(err)
}

func TestFormatHasEmptyPossibleRatings(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.PossibleRatings = []RatingId{}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected empty possible ratings to fail")
	}
	fmt.Println(err)
}

func TestFormatHasInvalidUserId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultFormat(t, db)

	format.UserId = database.InvalidUserId
	err := format.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Expected invalid user id to fail")
	}
	fmt.Println(err)
}

func TestFormatHasEmptyLines(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	format.Lines = []FormatLine{}
	err := format.StaticallyValid()
	if err == nil {
		t.Fatal("Expected empty lines to fail")
	}
	fmt.Println(err)
}

func TestFormatHasLineRatingsNotInPossibleLinesList(t *testing.T) {
	db := database.NewUnitTestDBProvider()
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
	db := database.NewUnitTestDBProvider()
	draft := newDefaultStoredDraft(t, db)

	_, _, err := database.DeleteOneById(context.Background(), db, &Format{}, draft.Format.RecordId())
	if err == nil {
		t.Fatal("Expected in-use format delete to fail")
	}
	fmt.Println(err)
}

func TestFormatCannotBeEditedWhenInUse(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	draft := newDefaultStoredDraft(t, db)

	format, err := database.GetExistingRecordById(context.Background(), db, &Format{}, draft.Format.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	newRating := newStoredRating(t, db)
	format.PossibleRatings = append(format.PossibleRatings, newRating.ID)
	err = database.UpdateOne(context.Background(), db, format)
	if err == nil {
		t.Fatal("Expected in-use format edit to fail")
	}
	fmt.Println(err)
}

func TestFormatCanBeEditedWhenNotInUse(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	f := newDefaultStoredFormat(t, db)

	newRating := newStoredRating(t, db)
	f.PossibleRatings = append(f.PossibleRatings, newRating.ID)

	err := database.UpdateOne(context.Background(), db, f)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormatPostCreateCreatesFormatRatings(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)

	formatRatings, err := database.GetAllWhere[*FormatRating](context.Background(), db, func(_ context.Context, fr *FormatRating) bool {
		return fr.FormatId == format.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(formatRatings) != len(format.PossibleRatings) {
		t.Fatalf("Expected %d FormatRating records, got %d", len(format.PossibleRatings), len(formatRatings))
	}
}

func TestFormatPostUpdateAddsNewRating(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	newRating := newStoredRating(t, db)

	format.PossibleRatings = append(format.PossibleRatings, newRating.ID)
	err := database.UpdateOne(context.Background(), db, format)
	if err != nil {
		t.Fatal(err)
	}

	formatRatings, err := database.GetAllWhere[*FormatRating](context.Background(), db, func(_ context.Context, fr *FormatRating) bool {
		return fr.FormatId == format.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(formatRatings) != len(format.PossibleRatings) {
		t.Fatalf("Expected %d FormatRating records, got %d", len(format.PossibleRatings), len(formatRatings))
	}

	found := false
	for _, fr := range formatRatings {
		if fr.RatingId == newRating.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Expected new rating to be in FormatRating records")
	}
}

func TestFormatPostUpdateRemovesOldRating(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	extraRating := newStoredRating(t, db)

	format.PossibleRatings = append(format.PossibleRatings, extraRating.ID)
	err := database.UpdateOne(context.Background(), db, format)
	if err != nil {
		t.Fatal(err)
	}

	formatRatings, err := database.GetAllWhere[*FormatRating](context.Background(), db, func(_ context.Context, fr *FormatRating) bool {
		return fr.FormatId == format.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(formatRatings) != 5 {
		t.Fatalf("Expected 5 FormatRating records after add, got %d", len(formatRatings))
	}

	format.PossibleRatings = format.PossibleRatings[:4]
	err = database.UpdateOne(context.Background(), db, format)
	if err != nil {
		t.Fatal(err)
	}

	formatRatings, err = database.GetAllWhere[*FormatRating](context.Background(), db, func(_ context.Context, fr *FormatRating) bool {
		return fr.FormatId == format.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(formatRatings) != 4 {
		t.Fatalf("Expected 4 FormatRating records after remove, got %d", len(formatRatings))
	}

	for _, fr := range formatRatings {
		if fr.RatingId == extraRating.ID {
			t.Fatal("Expected removed rating to not be in FormatRating records")
		}
	}
}

func TestFormatGetAssignedDrafts(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)
	user := newStoredUser(t, db)

	draft1 := NewDraft()
	draft1.Owner = user.ID
	draft1.Format = format.ID
	_, err := database.CreateOne(context.Background(), db, draft1)
	if err != nil {
		t.Fatal(err)

	}

	draft2 := NewDraft()
	draft2.Owner = user.ID
	draft2.Format = format.ID
	_, err = database.CreateOne(context.Background(), db, draft2)
	if err != nil {
		t.Fatal(err)
	}

	assigned, err := format.GetAssignedDrafts(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(assigned) != 2 {
		t.Fatalf("Expected 2 assigned drafts, got %d", len(assigned))
	}
}

func TestFormatGetAssignedDraftsEmpty(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)

	assigned, err := format.GetAssignedDrafts(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(assigned) != 0 {
		t.Fatalf("Expected 0 assigned drafts, got %d", len(assigned))
	}
}

func TestFormatRatingDuplicateFails(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)

	formatRatings, _ := database.GetAllWhere[*FormatRating](context.Background(), db, func(_ context.Context, fr *FormatRating) bool {
		return fr.FormatId == format.ID
	})

	duplicate := &FormatRating{
		FormatId: format.ID,
		RatingId: formatRatings[0].RatingId,
	}
	_, err := database.CreateOne(context.Background(), db, duplicate)
	if err == nil {
		t.Fatal("Expected duplicate FormatRating to fail")
	}
	fmt.Println(err)
}

func TestFormatRatingInvalidFormatId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	rating := newStoredRating(t, db)

	fr := &FormatRating{
		FormatId: FormatId(database.InvalidRecordId),
		RatingId: rating.ID,
	}
	err := fr.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Expected invalid format ID to fail validation")
	}
	fmt.Println(err)
}

func TestFormatRatingInvalidRatingId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)

	fr := &FormatRating{
		FormatId: format.ID,
		RatingId: RatingId(database.InvalidRecordId),
	}
	err := fr.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Expected invalid rating ID to fail validation")
	}
	fmt.Println(err)
}
