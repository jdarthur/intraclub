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
	r := NewRating()
	r.Name = "Rating 123"
	r.Description = "test description"
	v, err := common.CreateOne(db, r)
	if err != nil {
		t.Fatal(err)
	}
	return v
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
	t.Fatal("implement me")
}

func TestRatingCannotBeDeletedWhenInUse(t *testing.T) {
	t.Fatal("implement me")
}
