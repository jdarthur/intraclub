package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newValidRating(u database.UserId) *Rating {
	return &Rating{
		UserId:      u,
		Name:        "name",
		Description: "description",
	}
}

func newStoredRating(t *testing.T, db database.Provider) *Rating {
	user := newStoredUser(t, db)
	r := NewRating()
	r.UserId = user.ID
	r.Name = fmt.Sprintf("Rating %s", database.NewRecordId())
	r.Description = "test description"
	v, err := database.CreateOne(context.Background(), db, r)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func copyRating(r *Rating) *Rating {
	return &Rating{
		ID:          r.ID,
		UserId:      r.UserId,
		Name:        r.Name,
		Description: r.Description,
	}
}

func TestRatingNameEmpty(t *testing.T) {
	r := NewRating()
	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("expected error for empty rating name")
	}
	fmt.Println(err)
}

func TestRatingNameWhitespace(t *testing.T) {
	r := newValidRating(database.InvalidUserId)
	r.Name = "   "
	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("expected error for whitespace rating name")
	}
	fmt.Println(err)
}

func TestRatingDescriptionEmpty(t *testing.T) {
	r := newValidRating(database.InvalidUserId)
	r.Description = ""

	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("expected error for empty rating description")
	}
	fmt.Println(err)
}

func TestRatingDescriptionWhitespace(t *testing.T) {
	r := newValidRating(database.InvalidUserId)
	r.Description = "\n\n\n\n"

	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("expected error for whitespace rating description")
	}
	fmt.Println(err)
}

func TestRatingUserIdNotValid(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	r := newValidRating(database.InvalidUserId)

	err := r.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
	fmt.Println(err)
}

func TestRatingUpdateBySysAdmin(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	r := newStoredRating(t, db)
	sysAdmin := newSysAdmin(t, db)

	wac := database.WithAccessControl[*Rating]{Database: db, AccessControlUser: sysAdmin.ID}

	copied := copyRating(r)
	copied.Name = "new name"

	err := wac.UpdateOneById(context.Background(), copied)
	if err != nil {
		t.Fatal(err)
	}

	v, err := database.GetExistingRecordById(context.Background(), db, &Rating{}, r.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	if v.Name != copied.Name {
		t.Fatal("name not updated")
	}
}

func TestRatingCannotBeDeletedWhenInUse(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)

	possibleRatings, err := format.GetPossibleRatings(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	ratingId := possibleRatings[0].RecordId()
	rating, err := database.GetExistingRecordById(context.Background(), db, &Rating{}, ratingId)
	if err != nil {
		t.Fatal(err)
	}

	wac := database.WithAccessControl[*Rating]{Database: db, AccessControlUser: rating.UserId}
	_, _, err = wac.DeleteOneById(context.Background(), &Rating{}, ratingId)
	if err == nil {
		t.Fatal("Expected error on delete of in-use rating")
	}
	fmt.Println(err)

	_, exists, err := database.GetOneById(context.Background(), db, &Rating{}, ratingId)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Expected rating not to have been deleted")
	}

}

func TestRatingPreDeleteSucceedsWhenUnreferenced(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	rating := newStoredRating(t, db)

	err := rating.PreDelete(context.Background(), db)
	require.NoError(t, err)
}

func TestRatingPreDeleteBlockedByTeamRating(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	rating := newStoredRating(t, db)
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)

	_, err := database.CreateOne(context.Background(), db, &TeamRating{
		TeamId:   team.ID,
		UserId:   member.ID,
		RatingId: rating.ID,
	})
	require.NoError(t, err)

	err = rating.PreDelete(context.Background(), db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "team_rating")
}

func TestRatingPreDeleteBlockedByDraftRatingCutoff(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	rating := newStoredRating(t, db)
	draft := newDefaultStoredDraft(t, db)

	_, err := database.CreateOne(context.Background(), db, &DraftRatingCutoff{
		DraftId:     draft.ID,
		RatingId:    rating.ID,
		CutoffIndex: 1,
	})
	require.NoError(t, err)

	err = rating.PreDelete(context.Background(), db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "draft_rating_cutoff")
}

func TestRatingPreDeleteBlockedByDraftPick(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	rating := newStoredRating(t, db)
	draft := newDefaultStoredDraft(t, db)
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)

	_, err := database.CreateOne(context.Background(), db, &DraftPick{
		DraftId: draft.ID,
		TeamId:  team.ID,
		UserId:  member.ID,
		Round:   1,
		Pick:    1,
		Rating:  rating.ID,
	})
	require.NoError(t, err)

	err = rating.PreDelete(context.Background(), db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "draft_pick")
}

func TestRatingPreDeleteBlockedByPreDraftGrade(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	grade := newStoredGrade(t, db)
	rating, err := database.GetExistingRecordById(context.Background(), db, &Rating{}, grade.Rating.RecordId())
	require.NoError(t, err)

	err = rating.PreDelete(context.Background(), db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pre_draft_grade")
}

func TestRatingPreDeleteBlockedByFormatLine(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	rating := newStoredRating(t, db)
	user := newStoredUser(t, db)
	format := NewFormat()
	format.UserId = user.ID
	format.Name = "line test format"
	created, err := database.CreateOne(context.Background(), db, format)
	require.NoError(t, err)

	_, err = database.CreateOne(context.Background(), db, &FormatLine{
		FormatId:      created.ID,
		FormatIndex:   0,
		Player1Rating: rating.ID,
		Player2Rating: rating.ID,
	})
	require.NoError(t, err)

	err = rating.PreDelete(context.Background(), db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "format_line")
}
