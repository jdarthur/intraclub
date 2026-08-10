package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStoredCommissionerProposalForSeason(t *testing.T, db database.Provider, season *Season, mustBeUnanimous bool) *CommissionerProposal {
	proposal := NewCommissionerProposal()
	proposal.Description = "test description"
	proposal.SeasonId = season.ID
	proposal.MustBeUnanimous = mustBeUnanimous

	v, err := database.CreateOne(context.Background(), db, proposal)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newStoredCommissionerProposal(t *testing.T, db database.Provider, mustBeUnanimous bool) (*Season, *CommissionerProposal) {
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	proposal := newStoredCommissionerProposalForSeason(t, db, season, mustBeUnanimous)
	return season, proposal
}

func assertProposalStatus(t *testing.T, proposal *CommissionerProposal, db database.Provider, expectAccepted, expectRejected bool) {
	accepted, rejected, err := proposal.Status(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if accepted && !expectAccepted {
		t.Fatal("expected proposal to be not yet accepted")
	} else if !accepted && expectAccepted {
		t.Fatal("expected proposal to be accepted")
	} else if rejected && !expectRejected {
		t.Fatal("expected proposal to be not yet rejected")
	} else if !rejected && expectRejected {
		t.Fatal("expected proposal to be rejected")
	}
}

func TestCommissionerProposalUnanimousConsent(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, true)

	commissioners, err := season.GetCommissioners(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	for _, commissioner := range commissioners {
		err := prop.Vote(context.Background(), db, commissioner, true)
		if err != nil {
			t.Fatal(err)
		}
		assertProposalStatus(t, prop, db, false, false)
	}

	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	for i, team := range teams {
		captain, err := team.GetCaptain(context.Background(), db)
		if err != nil {
			t.Fatal(err)
		}
		err = prop.Vote(context.Background(), db, captain, true)
		if err != nil {
			t.Fatal(err)
		}
		if i < len(teams)-1 {
			// expect not yet accepted or rejected
			assertProposalStatus(t, prop, db, false, false)
		} else {
			// expect accepted after final "yes" vot
			assertProposalStatus(t, prop, db, true, false)
		}
	}
}

func TestCommissionerProposalUnanimousConsentOneNoRejects(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, true)

	commissioners, _ := season.GetCommissioners(context.Background(), db)
	err := prop.Vote(context.Background(), db, commissioners[0], false)
	if err != nil {
		t.Fatal(err)
	}
	assertProposalStatus(t, prop, db, false, true)
}

func TestCommissionerProposalFiftyPercentPlusOneAccepted(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, false)

	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	for _, team := range teams[:2] {
		captain, err := team.GetCaptain(context.Background(), db)
		if err != nil {
			t.Fatal(err)
		}
		err = prop.Vote(context.Background(), db, captain, true)
		if err != nil {
			t.Fatal(err)
		}
		assertProposalStatus(t, prop, db, false, false)
	}

	commissioners, _ := season.GetCommissioners(context.Background(), db)
	err = prop.Vote(context.Background(), db, commissioners[0], true)
	if err != nil {
		t.Fatal(err)
	}
	assertProposalStatus(t, prop, db, true, false)
}

func TestCommissionerProposalFiftyPercentPlusOneRejected(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, false)

	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	for _, team := range teams[:2] {
		captain, err := team.GetCaptain(context.Background(), db)
		if err != nil {
			t.Fatal(err)
		}
		err = prop.Vote(context.Background(), db, captain, false)
		if err != nil {
			t.Fatal(err)
		}
		assertProposalStatus(t, prop, db, false, false)
	}

	commissioners, _ := season.GetCommissioners(context.Background(), db)
	err = prop.Vote(context.Background(), db, commissioners[0], false)
	if err != nil {
		t.Fatal(err)
	}
	assertProposalStatus(t, prop, db, false, true)
}

func TestCommissionerProposalTieIsRejected(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 5)
	prop := newStoredCommissionerProposalForSeason(t, db, season, false)

	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	for _, team := range teams[:2] {
		captain, err := team.GetCaptain(context.Background(), db)
		if err != nil {
			t.Fatal(err)
		}
		err = prop.Vote(context.Background(), db, captain, false)
		if err != nil {
			t.Fatal(err)
		}
		assertProposalStatus(t, prop, db, false, false)
	}

	commissioners, _ := season.GetCommissioners(context.Background(), db)
	err = prop.Vote(context.Background(), db, commissioners[0], false)
	if err != nil {
		t.Fatal(err)
	}
	assertProposalStatus(t, prop, db, false, true)
}

func TestCommissionerProposalInvalidVoterId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, prop := newStoredCommissionerProposal(t, db, true)
	otherUser := newStoredUser(t, db)
	err := prop.Vote(context.Background(), db, otherUser.ID, false)
	if err == nil {
		t.Fatal("expected error on invalid voter")
	}
	fmt.Println(err)
}

func copyProposal(p *CommissionerProposal) *CommissionerProposal {
	return &CommissionerProposal{
		ID:              p.ID,
		Description:     p.Description,
		SeasonId:        p.SeasonId,
		MustBeUnanimous: p.MustBeUnanimous,
	}
}

func TestCommissionerProposalUnanimousConstraintCannotBeUpdated(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, prop := newStoredCommissionerProposal(t, db, true)
	copied := copyProposal(prop)
	copied.MustBeUnanimous = false
	err := database.UpdateOne(context.Background(), db, copied)
	if err == nil {
		t.Fatal("expected error when updating unanimous constraint")
	}
	fmt.Println(err)
}

func TestCommissionerProposalVoteRoundTrip(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, false)
	commissioners, err := season.GetCommissioners(context.Background(), db)
	require.NoError(t, err)

	row := &CommissionerProposalVote{
		ProposalId: prop.ID,
		UserId:     commissioners[0],
		Vote:       true,
	}
	created, err := database.CreateOne(context.Background(), db, row)
	require.NoError(t, err)

	got, exists, err := database.GetOneById(context.Background(), db, &CommissionerProposalVote{}, created.GetId())
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, prop.ID, got.ProposalId)
	assert.Equal(t, commissioners[0], got.UserId)
	assert.True(t, got.Vote)
}

func TestCommissionerProposalVoteUniquenessConstraint(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, false)
	commissioners, err := season.GetCommissioners(context.Background(), db)
	require.NoError(t, err)

	row1 := &CommissionerProposalVote{ProposalId: prop.ID, UserId: commissioners[0], Vote: true}
	_, err = database.CreateOne(context.Background(), db, row1)
	require.NoError(t, err)

	// a second vote for the same (proposal, user) must be rejected
	row2 := &CommissionerProposalVote{ProposalId: prop.ID, UserId: commissioners[0], Vote: false}
	_, err = database.CreateOne(context.Background(), db, row2)
	assert.Error(t, err)
}

func TestCommissionerProposalVoteDynamicallyValid(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, false)
	commissioners, err := season.GetCommissioners(context.Background(), db)
	require.NoError(t, err)

	valid := &CommissionerProposalVote{ProposalId: prop.ID, UserId: commissioners[0], Vote: true}
	assert.NoError(t, valid.DynamicallyValid(context.Background(), db))

	// invalid proposal ID
	badProp := &CommissionerProposalVote{ProposalId: database.InvalidRecordId, UserId: commissioners[0], Vote: true}
	assert.Error(t, badProp.DynamicallyValid(context.Background(), db))

	// invalid user ID
	badUser := &CommissionerProposalVote{ProposalId: prop.ID, UserId: database.InvalidUserId, Vote: true}
	assert.Error(t, badUser.DynamicallyValid(context.Background(), db))
}

func TestCommissionerProposalVoteViaVoteMethod(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, false)
	commissioners, err := season.GetCommissioners(context.Background(), db)
	require.NoError(t, err)

	err = prop.Vote(context.Background(), db, commissioners[0], true)
	require.NoError(t, err)

	votes, err := prop.GetVotes(context.Background(), db)
	require.NoError(t, err)
	assert.Len(t, votes, 1)
	assert.True(t, votes[commissioners[0]])
}

func TestCommissionerProposalVoteCanChange(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, prop := newStoredCommissionerProposal(t, db, false)
	commissioners, err := season.GetCommissioners(context.Background(), db)
	require.NoError(t, err)

	err = prop.Vote(context.Background(), db, commissioners[0], true)
	require.NoError(t, err)
	err = prop.Vote(context.Background(), db, commissioners[0], false)
	require.NoError(t, err)

	// the voter's existing relationship row is updated, not duplicated
	votes, err := prop.GetVotes(context.Background(), db)
	require.NoError(t, err)
	assert.Len(t, votes, 1)
	assert.False(t, votes[commissioners[0]])
}

func TestCommissionerProposalGetVotesEmpty(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, prop := newStoredCommissionerProposal(t, db, false)

	votes, err := prop.GetVotes(context.Background(), db)
	require.NoError(t, err)
	assert.Len(t, votes, 0)
}

func TestCommissionerProposalPostDeleteCascadesVotes(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, proposal := newStoredCommissionerProposal(t, db, false)

	voter := getAnyTeamCaptain(t, db, season)
	err := proposal.Vote(context.Background(), db, voter, true)
	require.NoError(t, err)

	count := func() int {
		rows, err := database.GetAllWhere[*CommissionerProposalVote](context.Background(), db, func(_ context.Context, v *CommissionerProposalVote) bool {
			return v.ProposalId == proposal.ID
		})
		require.NoError(t, err)
		return len(rows)
	}
	require.Equal(t, 1, count())

	_, _, err = database.DeleteOneById(context.Background(), db, &CommissionerProposal{}, proposal.ID)
	require.NoError(t, err)

	require.Equal(t, 0, count())
}
