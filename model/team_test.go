package model

import (
	"context"
	"testing"

	"intraclub/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStoredTeam(t *testing.T, db database.Provider, captain database.UserId) *Team {
	team := NewDefaultTeam(captain, "Test Team")

	v, err := database.CreateOne(context.Background(), db, team)
	if err != nil {
		t.Fatal(err)
	}

	// Create team assignment for captain
	assignment := &TeamAssignment{
		TeamId: v.ID,
		UserId: captain,
		Role:   TeamRoleCaptain,
	}
	_, err = database.CreateOne(context.Background(), db, assignment)
	if err != nil {
		t.Fatal(err)
	}

	return v
}

func TestTeamRatingRoundTrip(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)
	rating := newStoredRating(t, db)

	tr := &TeamRating{
		TeamId:   team.ID,
		UserId:   member.ID,
		RatingId: rating.ID,
	}
	created, err := database.CreateOne(context.Background(), db, tr)
	require.NoError(t, err)

	got, exists, err := database.GetOneById(context.Background(), db, &TeamRating{}, created.GetId())
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, team.ID, got.TeamId)
	assert.Equal(t, member.ID, got.UserId)
	assert.Equal(t, rating.ID, got.RatingId)
}

func TestTeamRatingUniquenessConstraint(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)

	rating1 := newStoredRating(t, db)
	rating2 := newStoredRating(t, db)

	tr1 := &TeamRating{TeamId: team.ID, UserId: member.ID, RatingId: rating1.ID}
	_, err := database.CreateOne(context.Background(), db, tr1)
	require.NoError(t, err)

	// a second TeamRating for the same (team, user) must be rejected
	tr2 := &TeamRating{TeamId: team.ID, UserId: member.ID, RatingId: rating2.ID}
	_, err = database.CreateOne(context.Background(), db, tr2)
	assert.Error(t, err)
}

func TestTeamRatingDynamicallyValid(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)
	rating := newStoredRating(t, db)

	// invalid team ID
	badTeam := &TeamRating{
		TeamId:   TeamId(database.InvalidRecordId),
		UserId:   member.ID,
		RatingId: rating.ID,
	}
	assert.Error(t, badTeam.DynamicallyValid(context.Background(), db))

	// invalid user ID
	badUser := &TeamRating{
		TeamId:   team.ID,
		UserId:   database.InvalidUserId,
		RatingId: rating.ID,
	}
	assert.Error(t, badUser.DynamicallyValid(context.Background(), db))

	// invalid rating ID
	badRating := &TeamRating{
		TeamId:   team.ID,
		UserId:   member.ID,
		RatingId: RatingId(database.InvalidRecordId),
	}
	assert.Error(t, badRating.DynamicallyValid(context.Background(), db))
}

func TestTeamGetRatingAndGetRatingsMap(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)
	rating := newStoredRating(t, db)

	tr := &TeamRating{TeamId: team.ID, UserId: member.ID, RatingId: rating.ID}
	_, err := database.CreateOne(context.Background(), db, tr)
	require.NoError(t, err)

	got, err := team.GetRating(context.Background(), db, member.ID)
	require.NoError(t, err)
	assert.Equal(t, rating.ID, got)

	// unknown user should error
	_, err = team.GetRating(context.Background(), db, newStoredUser(t, db).ID)
	assert.Error(t, err)

	ratingsMap, err := team.GetRatingsMap(context.Background(), db)
	require.NoError(t, err)
	assert.Equal(t, map[database.UserId]RatingId{member.ID: rating.ID}, ratingsMap)
}

func TestTeamGetRatingsMultipleMembers(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)

	member1 := newStoredUser(t, db)
	member2 := newStoredUser(t, db)
	rating1 := newStoredRating(t, db)
	rating2 := newStoredRating(t, db)

	_, err := database.CreateOne(context.Background(), db, &TeamRating{TeamId: team.ID, UserId: member1.ID, RatingId: rating1.ID})
	require.NoError(t, err)
	_, err = database.CreateOne(context.Background(), db, &TeamRating{TeamId: team.ID, UserId: member2.ID, RatingId: rating2.ID})
	require.NoError(t, err)

	ratingsMap, err := team.GetRatingsMap(context.Background(), db)
	require.NoError(t, err)
	assert.Equal(t, map[database.UserId]RatingId{
		member1.ID: rating1.ID,
		member2.ID: rating2.ID,
	}, ratingsMap)

	// each member resolves to their own rating
	got1, err := team.GetRating(context.Background(), db, member1.ID)
	require.NoError(t, err)
	assert.Equal(t, rating1.ID, got1)
	got2, err := team.GetRating(context.Background(), db, member2.ID)
	require.NoError(t, err)
	assert.Equal(t, rating2.ID, got2)
}

func TestTeamGetRatingsMapEmpty(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)

	ratingsMap, err := team.GetRatingsMap(context.Background(), db)
	require.NoError(t, err)
	assert.Empty(t, ratingsMap)
}

func TestTeamRatingQueryByTeam(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	otherTeam := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)
	otherMember := newStoredUser(t, db)
	rating := newStoredRating(t, db)

	_, err := database.CreateOne(context.Background(), db, &TeamRating{TeamId: team.ID, UserId: member.ID, RatingId: rating.ID})
	require.NoError(t, err)
	// a record for a different team should not appear in this team's query
	_, err = database.CreateOne(context.Background(), db, &TeamRating{TeamId: otherTeam.ID, UserId: otherMember.ID, RatingId: rating.ID})
	require.NoError(t, err)

	rows, err := database.GetAllWhere[*TeamRating](context.Background(), db, func(_ context.Context, r *TeamRating) bool {
		return r.TeamId == team.ID
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, member.ID, rows[0].UserId)
}

func TestTeamRatingUniquenessEquivalentDirect(t *testing.T) {
	teamId1 := TeamId(database.NewRecordId())
	teamId2 := TeamId(database.NewRecordId())
	user1 := database.UserId(database.NewRecordId())

	base := &TeamRating{TeamId: teamId1, UserId: user1}

	// same team + same user -> error
	assert.Error(t, base.UniquenessEquivalent(&TeamRating{TeamId: teamId1, UserId: user1}))

	// same user, different team -> allowed
	assert.NoError(t, base.UniquenessEquivalent(&TeamRating{TeamId: teamId2, UserId: user1}))

	// same team, different user -> allowed
	assert.NoError(t, base.UniquenessEquivalent(&TeamRating{TeamId: teamId1, UserId: database.UserId(database.NewRecordId())}))

	// different team + different user -> allowed
	assert.NoError(t, base.UniquenessEquivalent(&TeamRating{TeamId: teamId2, UserId: database.UserId(database.NewRecordId())}))
}

func TestTeamRatingDynamicallyValidSuccess(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)
	rating := newStoredRating(t, db)

	tr := &TeamRating{TeamId: team.ID, UserId: member.ID, RatingId: rating.ID}
	require.NoError(t, tr.DynamicallyValid(context.Background(), db))

	_, err := database.CreateOne(context.Background(), db, tr)
	require.NoError(t, err)
}

func TestTeamRatingUpdate(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member := newStoredUser(t, db)
	rating1 := newStoredRating(t, db)
	rating2 := newStoredRating(t, db)

	tr := &TeamRating{TeamId: team.ID, UserId: member.ID, RatingId: rating1.ID}
	created, err := database.CreateOne(context.Background(), db, tr)
	require.NoError(t, err)

	// reassigning the rating for the same (team, user) is a legal update
	created.RatingId = rating2.ID
	require.NoError(t, database.UpdateOne(context.Background(), db, created))

	got, err := team.GetRating(context.Background(), db, member.ID)
	require.NoError(t, err)
	assert.Equal(t, rating2.ID, got)
}

func TestTeamRatingUpdateToDuplicateRejected(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	team := newStoredTeam(t, db, newStoredUser(t, db).ID)
	member1 := newStoredUser(t, db)
	member2 := newStoredUser(t, db)
	rating := newStoredRating(t, db)

	tr1 := &TeamRating{TeamId: team.ID, UserId: member1.ID, RatingId: rating.ID}
	_, err := database.CreateOne(context.Background(), db, tr1)
	require.NoError(t, err)
	tr2 := &TeamRating{TeamId: team.ID, UserId: member2.ID, RatingId: rating.ID}
	created2, err := database.CreateOne(context.Background(), db, tr2)
	require.NoError(t, err)

	// moving member2 onto member1's (team, user) must be rejected
	created2.UserId = member1.ID
	assert.Error(t, database.UpdateOne(context.Background(), db, created2))
}

func TestTeamRatingAccessControl(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	tr := &TeamRating{}

	// ratings are owner-less, accessible to everyone, editable only by sysadmin
	assert.Equal(t, database.InvalidUserId, tr.GetOwner())
	assert.Equal(t, database.AccessibleToEveryone, tr.AccessibleTo(context.Background(), db))
	assert.Equal(t, []database.UserId{database.SysAdminUserId}, tr.EditableBy(context.Background(), db))

	// SetOwner is a no-op and does not change the owner
	tr.SetOwner(newStoredUser(t, db).ID)
	assert.Equal(t, database.InvalidUserId, tr.GetOwner())
}

func TestTeamRatingCrudRecordInterface(t *testing.T) {
	tr := &TeamRating{}

	assert.Equal(t, "team_rating", tr.Type())
	assert.NoError(t, tr.StaticallyValid())
	assert.Equal(t, database.RecordId(0), tr.GetId())

	// SetId/GetId round-trips
	tr.SetId(database.NewRecordId())
	assert.Equal(t, tr.ID, tr.GetId())

	// NewRecord returns a fresh empty record of the same type
	blank := tr.NewRecord()
	blankTr, ok := blank.(*TeamRating)
	require.True(t, ok)
	assert.Equal(t, database.RecordId(0), blankTr.GetId())
	assert.Equal(t, TeamId(0), blankTr.TeamId)
	assert.Equal(t, database.UserId(0), blankTr.UserId)
	assert.Equal(t, RatingId(0), blankTr.RatingId)
}

func TestTeamPostDeleteCascadesChildren(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	captain := newStoredUser(t, db)
	team := newStoredTeam(t, db, captain.ID) // creates a TeamAssignment

	member := newStoredUser(t, db)
	rating := newStoredRating(t, db)
	tr := &TeamRating{TeamId: team.ID, UserId: member.ID, RatingId: rating.ID}
	_, err := database.CreateOne(context.Background(), db, tr)
	require.NoError(t, err)

	counts := func() (int, int) {
		assignments, err := database.GetAllWhere[*TeamAssignment](context.Background(), db, func(_ context.Context, a *TeamAssignment) bool {
			return a.TeamId == team.ID
		})
		require.NoError(t, err)
		ratings, err := database.GetAllWhere[*TeamRating](context.Background(), db, func(_ context.Context, r *TeamRating) bool {
			return r.TeamId == team.ID
		})
		require.NoError(t, err)
		return len(assignments), len(ratings)
	}

	assignments, ratings := counts()
	require.Greater(t, assignments, 0)
	require.Greater(t, ratings, 0)

	_, _, err = database.DeleteOneById(context.Background(), db, &Team{}, team.ID.RecordId())
	require.NoError(t, err)

	assignments, ratings = counts()
	assert.Equal(t, 0, assignments)
	assert.Equal(t, 0, ratings)
}
