package model

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"intraclub/database"
)

func newValidBlurb(owner database.UserId, season SeasonId) *Blurb {
	b := NewBlurb()
	b.Owner = owner
	b.Season = season
	b.Title = "title"
	b.Content = "content"
	return b
}

func newDefaultBlurb(t *testing.T, db database.Provider) (*Blurb, *Season) {
	season, commissioner := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commissioner.ID, season.ID)
	return b, season
}

func newStoredBlurb(t *testing.T, db database.Provider, owner database.UserId, season SeasonId) *Blurb {
	b := newValidBlurb(owner, season)
	v, err := database.CreateOne(context.Background(), db, b)
	require.NoError(t, err)
	return v
}

func TestBlurbTitleIsEmpty(t *testing.T) {
	b := NewBlurb()
	assert.Error(t, b.StaticallyValid(), "expected error on empty title")
	fmt.Println(b.StaticallyValid())
}

func TestBlurbTitleIsOnlyWhitespace(t *testing.T) {
	b := NewBlurb()
	b.Title = " \n"
	assert.Error(t, b.StaticallyValid(), "expected error on whitespace title")
	fmt.Println(b.StaticallyValid())
}

func TestBlurbContentIsEmpty(t *testing.T) {
	b := NewBlurb()
	b.Title = "title"
	assert.Error(t, b.StaticallyValid(), "expected error on empty content")
	fmt.Println(b.StaticallyValid())
}

func TestBlurbContentIsOnlyWhitespace(t *testing.T) {
	b := NewBlurb()
	b.Title = "title"
	b.Content = "\t\r"
	assert.Error(t, b.StaticallyValid(), "expected error on whitespace content")
	fmt.Println(b.StaticallyValid())
}

func TestBlurbUserIdDoesNotExist(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	b := newValidBlurb(database.InvalidUserId, SeasonId(0))
	assert.Error(t, b.DynamicallyValid(context.Background(), db), "expected error on invalid user ID")
	fmt.Println(b.DynamicallyValid(context.Background(), db))
}

func TestBlurbSeasonIdDoesNotExist(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	b := newValidBlurb(user.ID, SeasonId(0))
	assert.Error(t, b.DynamicallyValid(context.Background(), db), "expected error on invalid user ID")
	fmt.Println(b.DynamicallyValid(context.Background(), db))
}

func TestBlurbByNonSeasonParticipant(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeason(t, db)
	otherUser := newStoredUser(t, db)

	b := newValidBlurb(otherUser.ID, season.ID)
	err := b.DynamicallyValid(context.Background(), db)
	assert.Error(t, err, "expected error on blurb by non-season participant")
	fmt.Println(err)
}

func TestBlurbPhotoIdDoesNotExist(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)
	// raw-insert a blurb_photo row referencing a nonexistent photo, bypassing
	// BlurbPhoto.DynamicallyValid, so Blurb.DynamicallyValid is what rejects it
	if _, err := db.Create(context.Background(), &BlurbPhoto{BlurbId: b.ID, PhotoId: 0}); err != nil {
		t.Fatal(err)
	}
	assert.Error(t, b.DynamicallyValid(context.Background(), db), "expected error on invalid photo ID")
	fmt.Println(b.DynamicallyValid(context.Background(), db))
}

func TestBlurbPhotoDoesNotBelongToUser(t *testing.T) {
	db := database.NewUnitTestDBProvider()

	b, _ := newDefaultBlurb(t, db)

	user2 := newStoredUser(t, db)
	photo := newStoredPhoto(t, db, user2.ID)

	if _, err := db.Create(context.Background(), &BlurbPhoto{BlurbId: b.ID, PhotoId: photo.ID}); err != nil {
		t.Fatal(err)
	}

	assert.Error(t, b.DynamicallyValid(context.Background(), db), "expected error on non-owned photo ID")
	fmt.Println(b.DynamicallyValid(context.Background(), db))
}

func TestInvalidReaction(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)
	assert.Error(t, b.React(context.Background(), db, commish.ID, reactionType(99999)), "expected error on invalid reaction")
	fmt.Println(b.React(context.Background(), db, commish.ID, reactionType(99999)))
}

func TestUserIdIsNotAMemberOfSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)

	otherUser := newStoredUser(t, db)
	assert.Error(t, b.React(context.Background(), db, otherUser.ID, ThumbsUp), "expected error on reaction from user who is not participating in season")
	fmt.Println(b.React(context.Background(), db, otherUser.ID, ThumbsUp))
}

func TestBlurbReactionByNonSeasonParticipant(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)

	otherUser := newStoredUser(t, db)
	row := &BlurbReaction{BlurbId: b.ID, UserId: otherUser.ID, ReactionType: ThumbsUp}
	err := row.DynamicallyValid(context.Background(), db)
	assert.Error(t, err, "expected error on reaction row from user who is not participating in season")
	fmt.Println(err)
}

func TestBlurbReactionBySeasonParticipant(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)

	row := &BlurbReaction{BlurbId: b.ID, UserId: commish.ID, ReactionType: ThumbsUp}
	require.NoError(t, row.DynamicallyValid(context.Background(), db))
}

func TestUserSuccessfulReact(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)
	require.NoError(t, b.React(context.Background(), db, commish.ID, ThumbsUp))
}

func TestDuplicateReaction(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)
	require.NoError(t, b.React(context.Background(), db, commish.ID, ThumbsUp))
	assert.Error(t, b.React(context.Background(), db, commish.ID, ThumbsUp), "expected error on duplicate reaction")
	fmt.Println(b.React(context.Background(), db, commish.ID, ThumbsUp))
}

func TestReactAndUnreact(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)
	require.NoError(t, b.React(context.Background(), db, commish.ID, ThumbsUp))
	require.NoError(t, b.Unreact(context.Background(), db, commish.ID, ThumbsUp))
}

func TestUnreactWhereNotPresent(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)
	b := newStoredBlurb(t, db, commish.ID, season.ID)

	assert.Error(t, b.Unreact(context.Background(), db, commish.ID, ThumbsUp), "expected error on unreact where existing reaction doesn't exist")
	fmt.Println(b.Unreact(context.Background(), db, commish.ID, ThumbsUp))
}

func TestBlurbPostDeleteCascadesChildren(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	blurb, season := newDefaultBlurb(t, db)
	participant := getAnyTeamCaptain(t, db, season)

	photo := newStoredPhoto(t, db, participant)
	_, err := database.CreateOne(context.Background(), db, &BlurbPhoto{BlurbId: blurb.ID, PhotoId: photo.ID})
	require.NoError(t, err)

	err = blurb.React(context.Background(), db, participant, ThumbsUp)
	require.NoError(t, err)

	counts := func() (int, int) {
		photos, err := database.GetAllWhere[*BlurbPhoto](context.Background(), db, func(_ context.Context, p *BlurbPhoto) bool {
			return p.BlurbId == blurb.ID
		})
		require.NoError(t, err)
		reactions, err := database.GetAllWhere[*BlurbReaction](context.Background(), db, func(_ context.Context, r *BlurbReaction) bool {
			return r.BlurbId == blurb.ID
		})
		require.NoError(t, err)
		return len(photos), len(reactions)
	}

	photos, reactions := counts()
	require.Equal(t, 1, photos)
	require.Equal(t, 1, reactions)

	_, _, err = database.DeleteOneById(context.Background(), db, &Blurb{}, blurb.ID.RecordId())
	require.NoError(t, err)

	photos, reactions = counts()
	require.Equal(t, 0, photos)
	require.Equal(t, 0, reactions)
}
