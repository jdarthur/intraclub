package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newReaction(t *testing.T, db common.DatabaseProvider) *Reaction {
	user := newStoredUser(t, db)
	reaction := &Reaction{
		UserId: user.ID,
		Type:   ThumbsUp,
	}
	return reaction
}

func TestReactionAlreadyPresent(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	reaction := newReaction(t, db)

	r := make(ReactionList, 0)
	r = append(r, reaction)
	err := r.DynamicallyValid(db)
	if err != nil {
		t.Fatal(err)
	}

	err = r.CanAddReaction(db, reaction)
	if err == nil {
		t.Fatal("should get error re-adding an existing reaction")
	}
	fmt.Println(err)
}

func TestDuplicateReactionNotStaticallyValid(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	reaction := newReaction(t, db)
	reaction2 := newReaction(t, db)

	r := make(ReactionList, 0)
	r = append(r, reaction, reaction2, reaction)

	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("should get error for a duplicate reaction")
	}
	fmt.Println(err)
}

func TestReactionUserIdDoesNotExist(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	reaction := &Reaction{
		UserId: 0,
		Type:   ThumbsUp,
	}

	r := make(ReactionList, 0)
	r = append(r, reaction)

	err := r.DynamicallyValid(db)
	if err == nil {
		t.Fatal("should get error for a nonexistent user ID in reaction")
	}
	fmt.Println(err)
}

func TestReactionUserTypeIsInvalid(t *testing.T) {
	reaction := &Reaction{
		UserId: 0,
		Type:   reactionType(999999),
	}

	r := make(ReactionList, 0)
	r = append(r, reaction)

	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("should get error for an invalid reaction type")
	}
	fmt.Println(err)
}
