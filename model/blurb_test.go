package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newValidBlurb(owner UserId, season SeasonId) *Blurb {
	b := NewBlurb()
	b.Owner = owner
	b.Season = season
	b.Title = "title"
	b.Content = "content"
	return b
}

func newDefaultBlurb(t *testing.T, db common.DatabaseProvider) (*Blurb, *Season) {
	season := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, season.Commissioners[0], season.ID)
	return b, season
}

func newStoredBlurb(t *testing.T, db common.DatabaseProvider, owner UserId, season SeasonId) *Blurb {
	b := newValidBlurb(owner, season)
	v, err := common.CreateOne(db, b)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestBlurbTitleIsEmpty(t *testing.T) {
	b := NewBlurb()
	err := b.StaticallyValid()
	if err == nil {
		t.Fatal("expected error on empty title")
	}
	fmt.Println(err)
}

func TestBlurbTitleIsOnlyWhitespace(t *testing.T) {
	b := NewBlurb()
	b.Title = " \n"
	err := b.StaticallyValid()
	if err == nil {
		t.Fatal("expected error on whitespace title")
	}
	fmt.Println(err)
}

func TestBlurbContentIsEmpty(t *testing.T) {
	b := NewBlurb()
	b.Title = "title"
	err := b.StaticallyValid()
	if err == nil {
		t.Fatal("expected error on empty content")
	}
	fmt.Println(err)
}

func TestBlurbContentIsOnlyWhitespace(t *testing.T) {
	b := NewBlurb()
	b.Title = "title"
	b.Content = "\t\r"
	err := b.StaticallyValid()
	if err == nil {
		t.Fatal("expected error on whitespace content")
	}
	fmt.Println(err)
}

func TestBlurbUserIdDoesNotExist(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	b := newValidBlurb(UserId(common.InvalidRecordId), SeasonId(0))
	err := b.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected error on invalid user ID")
	}
	fmt.Println(err)
}

func TestBlurbSeasonIdDoesNotExist(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	b := newValidBlurb(user.ID, SeasonId(0))
	err := b.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected error on invalid user ID")
	}
	fmt.Println(err)
}

func TestBlurbPhotoIdDoesNotExist(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	b := newValidBlurb(season.Commissioners[0], season.ID)
	b.Photos = []PhotoId{0}
	err := b.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected error on invalid photo ID")
	}
	fmt.Println(err)
}

func TestBlurbPhotoDoesNotBelongToUser(t *testing.T) {
	db := common.NewUnitTestDBProvider()

	b, _ := newDefaultBlurb(t, db)

	user2 := newStoredUser(t, db)
	photo := newStoredPhoto(t, db, user2.ID)

	b.Photos = []PhotoId{photo.ID}

	err := b.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected error on non-owned photo ID")
	}
	fmt.Println(err)
}

func TestInvalidReaction(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, season.Commissioners[0], season.ID)
	err := b.React(db, season.Commissioners[0], reactionType(99999))
	if err == nil {
		t.Fatal("expected error on invalid reaction")
	}
	fmt.Println(err)
}

func TestUserIdIsNotAMemberOfSeason(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	otherUser := newStoredUser(t, db)

	b := newStoredBlurb(t, db, season.Commissioners[0], season.ID)
	err := b.React(db, otherUser.ID, ThumbsUp)
	if err == nil {
		t.Fatal("expected error on reaction from user who is not participating in season")
	}
	fmt.Println(err)
}

func TestUserSuccessfulReact(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, season.Commissioners[0], season.ID)
	err := b.React(db, season.Commissioners[0], ThumbsUp)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDuplicateReaction(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, season.Commissioners[0], season.ID)
	err := b.React(db, season.Commissioners[0], ThumbsUp)
	if err != nil {
		t.Fatal(err)
	}
	err = b.React(db, season.Commissioners[0], ThumbsUp)
	if err == nil {
		t.Fatal("expected error on duplicate reaction")
	}
	fmt.Println(err)
}

func TestReactAndUnreact(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, season.Commissioners[0], season.ID)
	err := b.React(db, season.Commissioners[0], ThumbsUp)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Unreact(db, season.Commissioners[0], ThumbsUp)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnreactWhereNotPresent(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, season.Commissioners[0], season.ID)
	err := b.Unreact(db, season.Commissioners[0], ThumbsUp)
	if err == nil {
		t.Fatal("expected error on unreact where existing reaction doesn't exist")
	}
	fmt.Println(err)
}
