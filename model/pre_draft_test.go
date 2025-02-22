package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

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
