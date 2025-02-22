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

	f.Lines = lines

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

func TestGetFormatAvailableRatings(t *testing.T) {
	db := common.NewUnitTestDBProvider()

	line1 := newLine(t, db)
	line2 := newLine(t, db)
	line3 := Line{
		Player1Rating: line1.Player1Rating,
		Player2Rating: line2.Player1Rating,
	}
	format := NewFormat()
	format.Lines = []Line{line1, line2, line3}

	availableRatings := format.GetAvailableRatings()
	if len(availableRatings) != 4 {
		t.Fatalf("Expected 3 ratings, got %d", len(availableRatings))
	}
	if availableRatings[0] != line1.Player1Rating {
		t.Fatalf("Expected first rating to be line1.Player1Rating, got %s", availableRatings[0])
	}
	if availableRatings[1] != line1.Player2Rating {
		t.Fatalf("Expected first rating to be line1.Player2Rating, got %s", availableRatings[1])
	}
	if availableRatings[2] != line2.Player1Rating {
		t.Fatalf("Expected first rating to be line2.Player1Rating, got %s", availableRatings[2])
	}
	if availableRatings[3] != line2.Player2Rating {
		t.Fatalf("Expected first rating to be line3.Player2Rating, got %s", availableRatings[3])
	}
}
