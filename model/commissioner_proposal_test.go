package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredCommissionerProposalForSeason(t *testing.T, db database.DatabaseProvider, season *Season, mustBeUnanimous bool) *CommissionerProposal {
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

func newStoredCommissionerProposal(t *testing.T, db database.DatabaseProvider, mustBeUnanimous bool) (*Season, *CommissionerProposal) {
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	proposal := newStoredCommissionerProposalForSeason(t, db, season, mustBeUnanimous)
	return season, proposal
}

func assertProposalStatus(t *testing.T, proposal *CommissionerProposal, db database.DatabaseProvider, expectAccepted, expectRejected bool) {
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
		Votes:           p.Votes,
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
