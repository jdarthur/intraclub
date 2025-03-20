package model

import (
	"fmt"
	"intraclub/common"
	"math/rand"
	"testing"
)

func generateRandomPreDraftGrades(t *testing.T, db common.DatabaseProvider, playerCount, teamCount int) *Draft {
	randomDraft := newRandomDraft(t, db, playerCount, teamCount)
	gradeCount := 10
	if playerCount < 10 {
		gradeCount = playerCount
	}

	format, err := common.GetExistingRecordById(db, &Format{}, randomDraft.Format.RecordId())
	if err != nil {
		t.Fatal(err)
	}

	for _, playerId := range randomDraft.Available {

		// create X grades per player
		for i := 0; i < gradeCount; i++ {
			// grader will just be the first X user IDs in the available list
			grader := randomDraft.Available[i]

			ratingIndex := rand.Intn(len(format.PossibleRatings)) // random rating
			modifier := PreDraftRatingModifier(rand.Intn(3))      // random modifier

			grade := NewPreDraftGrade()
			grade.PlayerId = playerId
			grade.GraderId = grader
			grade.DraftId = randomDraft.ID
			grade.Rating = format.PossibleRatings[ratingIndex]
			grade.Modifier = modifier

			// create the grade
			_, err := common.CreateOne(db, grade)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	return randomDraft
}

func newValidGrade(t *testing.T, db common.DatabaseProvider) *PreDraftGrade {
	randomDraft := newRandomDraft(t, db, 4, 4)
	format, err := common.GetExistingRecordById(db, &Format{}, randomDraft.Format.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	grade := NewPreDraftGrade()
	grade.PlayerId = randomDraft.Available[0]
	grade.GraderId = randomDraft.Available[0]
	grade.DraftId = randomDraft.ID
	grade.Rating = format.PossibleRatings[0]
	return grade
}

func TestPreDraftInvalidRating(t *testing.T) {

	db := common.NewUnitTestDBProvider()
	rating := newStoredRating(t, db)
	draft := newRandomDraft(t, db, 5, 2)

	player := draft.Available[0]
	grader := draft.Available[1]

	grade := NewPreDraftGrade()
	grade.PlayerId = player
	grade.GraderId = grader
	grade.DraftId = draft.ID
	grade.Rating = rating.ID

	err := grade.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}

func TestPreDraftInvalidPlayerId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	grade := newValidGrade(t, db)
	grade.PlayerId = UserId(common.InvalidRecordId)
	err := grade.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}

func TestPreDraftInvalidGraderId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	grade := newValidGrade(t, db)
	grade.GraderId = UserId(common.InvalidRecordId)
	err := grade.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}

func TestPreDraftInvalidDraftId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	grade := newValidGrade(t, db)
	grade.DraftId = DraftId(common.InvalidRecordId)
	err := grade.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}

func TestPreDraftInvalidRatingId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	grade := newValidGrade(t, db)
	grade.Rating = RatingId(common.InvalidRecordId)
	err := grade.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}

func TestPreDraftPlayerIsNotAvailableToDraft(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	grade := newValidGrade(t, db)
	someOtherUser := newStoredUser(t, db)
	grade.PlayerId = someOtherUser.ID
	err := grade.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}

func TestPreDraftRatingIsNotPresentInDraftFormat(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	grade := newValidGrade(t, db)
	someOtherRating := newStoredRating(t, db)
	grade.Rating = someOtherRating.ID
	err := grade.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}

func TestModifierIsOutsideRange(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	grade := newValidGrade(t, db)
	grade.Modifier = PreDraftRatingModifier(-1)
	err := grade.StaticallyValid()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}

func TestGetRatingAggregate(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := generateRandomPreDraftGrades(t, db, 20, 4)

	aggregates, err := GetSortedListOfAllPreDraftGradesDescending(db, draft)
	if err != nil {
		t.Fatal(err)
	}
	for _, aggregate := range aggregates {
		fmt.Printf("%+v\n", aggregate)
	}
}

func TestCalculateNumericGrades(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 4, 4)
	playerToGrade := draft.Available[0]
	format, err := common.GetExistingRecordById(db, &Format{}, draft.Format.RecordId())
	if err != nil {
		t.Fatal(err)
	}

	grade := NewPreDraftGrade()
	grade.PlayerId = playerToGrade
	grade.GraderId = playerToGrade
	grade.DraftId = draft.ID
	grade.Rating = format.PossibleRatings[0]

	numeric := grade.NumericRating(format)
	if numeric != 10 {
		t.Fatalf("Expected 10 numeric rating, got %f", numeric)
	}

	grade.Rating = format.PossibleRatings[1]
	numeric = grade.NumericRating(format)
	if numeric != 7 {
		t.Fatalf("Expected 7 numeric rating, got %f", numeric)
	}

	grade.Rating = format.PossibleRatings[2]
	numeric = grade.NumericRating(format)
	if numeric != 4 {
		t.Fatalf("Expected 4 numeric rating, got %f", numeric)
	}

	grade.Rating = format.PossibleRatings[3]
	numeric = grade.NumericRating(format)
	if numeric != 1 {
		t.Fatalf("Expected 1 numeric rating, got %f", numeric)
	}
}

func TestGradeWhenDraftIsCompleted(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := newRandomDraft(t, db, 20, 4)
	format, err := common.GetExistingRecordById(db, &Format{}, draft.Format.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	completeExistingDraft(t, draft)

	grade := NewPreDraftGrade()
	grade.PlayerId = draft.Available[0]
	grade.GraderId = draft.Available[0]
	grade.DraftId = draft.ID
	grade.Rating = format.PossibleRatings[0]

	err = grade.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	fmt.Println(err)
}
