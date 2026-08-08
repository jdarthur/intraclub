package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

// newLine creates a FormatLine referencing two stored ratings.
func newLine(t *testing.T, db database.Provider) FormatLine {
	return FormatLine{
		Player1Rating: newStoredRating(t, db).ID,
		Player2Rating: newStoredRating(t, db).ID,
	}
}

// appendUniqueRating appends r to ratings unless it is already present.
func appendUniqueRating(ratings RatingList, r RatingId) RatingList {
	for _, existing := range ratings {
		if existing == r {
			return ratings
		}
	}
	return append(ratings, r)
}

// newDefaultFormat creates and stores a format with two lines, whose unique
// ratings become the format's possible ratings (stored in the format_rating /
// format_line join tables).
func newDefaultFormat(t *testing.T, db database.Provider) *Format {
	user := newStoredUser(t, db)
	f := NewFormat()
	f.UserId = user.ID
	f.Name = "default format"

	created, err := database.CreateOne(context.Background(), db, f)
	if err != nil {
		t.Fatal(err)
	}

	lines := []FormatLine{
		newLine(t, db),
		newLine(t, db),
	}
	var ratings RatingList
	for _, l := range lines {
		ratings = appendUniqueRating(ratings, l.Player1Rating)
		ratings = appendUniqueRating(ratings, l.Player2Rating)
	}
	if err := created.SetPossibleRatings(context.Background(), db, ratings); err != nil {
		t.Fatal(err)
	}
	if err := created.SetLines(context.Background(), db, lines); err != nil {
		t.Fatal(err)
	}
	return created
}

func newDefaultStoredFormat(t *testing.T, db database.Provider) *Format {
	return newDefaultFormat(t, db)
}

func TestFormatDuplicateLine(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ctx := context.Background()

	format := newDefaultStoredFormat(t, db)
	lines, err := format.GetLines(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	dup := lines[0]

	err = format.SetLines(ctx, db, []FormatLine{dup, dup})
	if err == nil {
		t.Fatal("Expected duplicate line to fail")
	}
	fmt.Println(err)
}

func TestFormatReversedValueDuplicateLine(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ctx := context.Background()

	format := newDefaultStoredFormat(t, db)
	lines, err := format.GetLines(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	line1 := lines[0]
	line2 := FormatLine{Player1Rating: line1.Player2Rating, Player2Rating: line1.Player1Rating}

	err = format.SetLines(ctx, db, []FormatLine{line1, line2})
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

	err := format.SetPossibleRatings(context.Background(), db, nil)
	if err == nil {
		t.Fatal("Expected empty possible ratings to fail")
	}
	fmt.Println(err)
}

func TestFormatHasInvalidUserId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)

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

	err := format.SetLines(context.Background(), db, nil)
	if err == nil {
		t.Fatal("Expected empty lines to fail")
	}
	fmt.Println(err)
}

func TestFormatHasLineRatingsNotInPossibleLinesList(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)

	other := newStoredRating(t, db)
	err := format.SetLines(context.Background(), db, []FormatLine{
		{Player1Rating: other.ID, Player2Rating: other.ID},
	})
	if err == nil {
		t.Fatal("Expected line ratings not in possible list to fail")
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
	err = database.UpdateOne(context.Background(), db, format)
	if err == nil {
		t.Fatal("Expected in-use format edit to fail")
	}
	fmt.Println(err)
}

func TestFormatCanBeEditedWhenNotInUse(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	f := newDefaultStoredFormat(t, db)

	f.Name = "renamed format"
	err := database.UpdateOne(context.Background(), db, f)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFormatSetPossibleRatingsCreatesFormatRatings(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ctx := context.Background()

	format := newDefaultFormat(t, db)
	got, err := format.GetPossibleRatings(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	formatRatings, err := database.GetAllWhere[*FormatRating](ctx, db, func(_ context.Context, fr *FormatRating) bool {
		return fr.FormatId == format.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(formatRatings) != len(got) {
		t.Fatalf("Expected %d FormatRating records, got %d", len(got), len(formatRatings))
	}
}

func TestFormatSetPossibleRatingsAddsNewRating(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ctx := context.Background()

	format := newDefaultStoredFormat(t, db)
	before, err := format.GetPossibleRatings(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	newRating := newStoredRating(t, db)
	after := append(append(RatingList{}, before...), newRating.ID)
	if err := format.SetPossibleRatings(ctx, db, after); err != nil {
		t.Fatal(err)
	}

	got, err := format.GetPossibleRatings(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(after) {
		t.Fatalf("Expected %d FormatRating records, got %d", len(after), len(got))
	}

	found := false
	for _, r := range got {
		if r == newRating.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Expected new rating to be in FormatRating records")
	}
}

func TestFormatSetPossibleRatingsRemovesOldRating(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ctx := context.Background()

	format := newDefaultStoredFormat(t, db)
	before, err := format.GetPossibleRatings(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	extraRating := newStoredRating(t, db)
	withExtra := append(append(RatingList{}, before...), extraRating.ID)
	if err := format.SetPossibleRatings(ctx, db, withExtra); err != nil {
		t.Fatal(err)
	}

	withoutExtra := withExtra[:len(withExtra)-1]
	if err := format.SetPossibleRatings(ctx, db, withoutExtra); err != nil {
		t.Fatal(err)
	}

	got, err := format.GetPossibleRatings(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(withoutExtra) {
		t.Fatalf("Expected %d FormatRating records after remove, got %d", len(withoutExtra), len(got))
	}

	for _, r := range got {
		if r == extraRating.ID {
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
		FormatId:    format.ID,
		RatingId:    formatRatings[0].RatingId,
		RatingIndex: formatRatings[0].RatingIndex,
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
