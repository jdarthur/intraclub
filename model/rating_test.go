package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
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

	ratingId := format.PossibleRatings[0].RecordId()
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
