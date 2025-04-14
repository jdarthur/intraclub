package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"strings"
)

// CommissionerProposal is a type that allows a Season commissioner to propose a
// one-time (or perhaps permanent going-forward) rule change or other type of
// administrative action during the season (such as adding a player to a team
// after-the-fact or modifying a player's rating). This can be ratified either by
// majority rule (50%+1), or by unanimous consent, based on the type of proposal
type CommissionerProposal struct {
	ID              common.RecordId // unique ID for this proposal
	Description     string          // description of the change or action
	SeasonId        SeasonId        // season that this pertains to
	Votes           map[UserId]bool // votes of all commissioners or team captains, true == vote in favor, false == vote against
	MustBeUnanimous bool            // true if this proposal must get unanimous consent to pass
}

func (c *CommissionerProposal) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

func (c *CommissionerProposal) PreUpdate(db common.DatabaseProvider, existingValues common.CrudRecord) error {
	old := existingValues.(*CommissionerProposal)
	if c.MustBeUnanimous != old.MustBeUnanimous {
		return fmt.Errorf("'must be unanimous' constraint can not be updated after creation")
	}
	return nil
}

func NewCommissionerProposal() *CommissionerProposal {
	return &CommissionerProposal{
		Votes: make(map[UserId]bool),
	}
}

func (c *CommissionerProposal) Type() string {
	return "commissioner_proposal"
}

func (c *CommissionerProposal) GetId() common.RecordId {
	return c.ID
}

func (c *CommissionerProposal) SetId(id common.RecordId) {
	c.ID = id
}

func (c *CommissionerProposal) EditableBy(db common.DatabaseProvider) []common.RecordId {
	// proposal is editable only by the
	return EditableBySeason(db, c.SeasonId)
}

func (c *CommissionerProposal) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	// a commissioner proposal is only accessible to the other commissioners and the
	// team captains involved in the given Season.
	voters, err := c.GetAllVoterIds(db)
	if err != nil {
		fmt.Printf("Failed to get voters for commissioner proposal: %s\n", err.Error())
	}
	return UserIdListToRecordIdList(voters)
}

func (c *CommissionerProposal) SetOwner(recordId common.RecordId) {
	// don't need to do anything here as the ownership of the
	// CommissionerProposal record type is automatically inferred &
	// enforced by the associated Season assigned to it
}

func (c *CommissionerProposal) StaticallyValid() error {
	c.Description = strings.TrimSpace(c.Description)
	if c.Description == "" {
		return errors.New("empty description")
	}
	return nil
}

func (c *CommissionerProposal) DynamicallyValid(db common.DatabaseProvider) error {
	// this will return an error if the SeasonId on this proposal is not valid,
	// so we don't need to check for that scenario again in this function
	possibleVoters, err := c.GetAllVoterIds(db)
	if err != nil {
		return err
	}

	// validate that each voter is a valid captain or commissioner
	for voter, _ := range c.Votes {
		found := false
		for _, captain := range possibleVoters {
			if voter == captain {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("voter %s not found in possible voters for commissioner proposal", voter)
		}
	}
	return nil
}

func (c *CommissionerProposal) GetAllVoterIds(db common.DatabaseProvider) ([]UserId, error) {
	// get underlying season
	season, err := common.GetExistingRecordById(db, &Season{}, c.SeasonId.RecordId())
	if err != nil {
		return nil, err
	}
	// add all commissioners as valid voters
	output := make([]UserId, 0)
	output = append(output, season.Commissioners...)

	// get all team captains as the other voters
	teamCaptains, err := season.GetTeamCaptains(db)
	if err != nil {
		return nil, err
	}
	return append(output, teamCaptains...), nil
}

func (c *CommissionerProposal) Vote(db common.DatabaseProvider, voterId UserId, vote bool) error {
	// add vote to the map
	c.Votes[voterId] = vote

	// update the record, returning an error if e.g. this UserId is not entitled to a vote here
	return common.UpdateOne(db, c)
}

func (c *CommissionerProposal) VotesToPassOrFail(voterIds []UserId) (votesToPass, votesToFail int) {
	if c.MustBeUnanimous {
		// if the vote must be unanimous, then we need all voters to vote yes
		// to pass, and one voter to vote not to fail
		return len(voterIds), 1
	} else {
		// if the vote must not be yes, we need 50% +1 to pass or 50% to fail.
		// A tie vote in this scenario will be
		if len(voterIds)%2 == 1 {
			return (len(voterIds) / 2) + 1, len(voterIds)/2 + 1
		} else {
			return (len(voterIds) / 2) + 1, len(voterIds) / 2
		}
	}
}

func (c *CommissionerProposal) Status(db common.DatabaseProvider) (accepted, rejected bool, err error) {
	voterIds, err := c.GetAllVoterIds(db)
	if err != nil {
		return false, false, err
	}

	// X votes to pass, Y votes to fail based on
	votesNeededToPass, votesNeededToFail := c.VotesToPassOrFail(voterIds)

	votesInFavor := 0
	votesAgainst := 0
	for _, vote := range c.Votes {
		// tally all the existing votes

		if vote == true {
			votesInFavor += 1
		} else {
			votesAgainst += 1
		}

		// if we reach the pass threshold, then return
		if votesInFavor >= votesNeededToPass {
			return true, false, nil
		}
		// if we reach the fail threshold then return
		if votesAgainst >= votesNeededToFail {
			return false, true, nil
		}
	}
	// if we don't have enough votes to accept or reject
	// the proposal, then return that information back
	return false, false, nil
}
