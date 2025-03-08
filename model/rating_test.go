package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newValidRating(u UserId) *Rating {
	return &Rating{
		UserId:      u,
		Name:        "name",
		Description: "description",
	}
}

func newStoredRating(t *testing.T, db common.DatabaseProvider) *Rating {
	user := newStoredUser(t, db)
	r := NewRating()
	r.UserId = user.ID
	r.Name = "Rating 123"
	r.Description = "test description"
	v, err := common.CreateOne(db, r)
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
	r := newValidRating(UserId(common.InvalidRecordId))
	r.Name = "   "
	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("expected error for whitespace rating name")
	}
	fmt.Println(err)
}

func TestRatingDescriptionEmpty(t *testing.T) {
	r := newValidRating(UserId(common.InvalidRecordId))
	r.Description = ""

	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("expected error for empty rating description")
	}
	fmt.Println(err)
}

func TestRatingDescriptionWhitespace(t *testing.T) {
	r := newValidRating(UserId(common.InvalidRecordId))
	r.Description = "\n\n\n\n"

	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("expected error for whitespace rating description")
	}
	fmt.Println(err)
}

func TestRatingUserIdNotValid(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	r := newValidRating(UserId(common.InvalidRecordId))

	err := r.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
	fmt.Println(err)
}

func TestRatingUpdateBySysAdmin(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	r := newStoredRating(t, db)
	sysAdmin := newSysAdmin(t, db)

	wac := common.WithAccessControl[*Rating]{Database: db, AccessControlUser: sysAdmin.ID.RecordId()}

	copied := copyRating(r)
	copied.Name = "new name"

	err := wac.UpdateOneById(copied)
	if err != nil {
		t.Fatal(err)
	}

	v, err := common.GetExistingRecordById(db, &Rating{}, r.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}
	if v.Name != copied.Name {
		t.Fatal("name not updated")
	}
}

func TestRatingCannotBeDeletedWhenInUse(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	format := newDefaultStoredFormat(t, db)

	ratingId := format.PossibleRatings[0].RecordId()
	rating, err := common.GetExistingRecordById(db, &Rating{}, ratingId)
	if err != nil {
		t.Fatal(err)
	}

	wac := common.WithAccessControl[*Rating]{Database: db, AccessControlUser: rating.UserId.RecordId()}
	_, _, err = wac.DeleteOneById(&Rating{}, ratingId)
	if err == nil {
		t.Fatal("Expected error on delete of in-use rating")
	}
	fmt.Println(err)

	_, exists, err := common.GetOneById(db, &Rating{}, ratingId)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Expected rating not to have been deleted")
	}

}
