package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"intraclub/database"
)

// CommissionerProposal is a type that allows a Season commissioner to propose a
// one-time (or perhaps permanent going-forward) rule change or other type of
// administrative action during the season (such as adding a player to a team
// after-the-fact or modifying a player's rating). This can be ratified either by
// majority rule (50%+1), or by unanimous consent, based on the type of proposal.
//
// API/JSON shape decision: the former inline `votes` field
// (`map[database.UserId]bool`) has been removed from `CommissionerProposal` and
// normalized into the `CommissionerProposalVote` relationship table. A proposal
// is not currently exposed via a REST CRUD route (see main.go), so removing the
// field is not a breaking wire change; in-process reads can reassemble the
// relationship rows into the old map shape via `CommissionerProposal.GetVotes`.
type CommissionerProposal struct {
	ID              database.RecordId // unique ID for this proposal
	Description     string            // description of the change or action
	SeasonId        SeasonId          // season that this pertains to
	MustBeUnanimous bool              // true if this proposal must get unanimous consent to pass
}

func (c *CommissionerProposal) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (c *CommissionerProposal) PreUpdate(ctx context.Context, db database.Provider, existingValues database.CrudRecord) error {
	old := existingValues.(*CommissionerProposal)
	if c.MustBeUnanimous != old.MustBeUnanimous {
		return fmt.Errorf("'must be unanimous' constraint can not be updated after creation")
	}
	return nil
}

func NewCommissionerProposal() *CommissionerProposal {
	return &CommissionerProposal{}
}

func (c *CommissionerProposal) Type() string {
	return "commissioner_proposal"
}

func (c *CommissionerProposal) GetId() database.RecordId {
	return c.ID
}

func (c *CommissionerProposal) SetId(id database.RecordId) {
	c.ID = id
}

func (c *CommissionerProposal) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	// proposal is editable only by the
	return EditableBySeason(ctx, db, c.SeasonId)
}

func (c *CommissionerProposal) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	// a commissioner proposal is only accessible to the other commissioners and the
	// team captains involved in the given Season.
	voters, err := c.GetAllVoterIds(ctx, db)
	if err != nil {
		fmt.Printf("Failed to get voters for commissioner proposal: %s\n", err.Error())
	}
	return voters
}

func (c *CommissionerProposal) SetOwner(userId database.UserId) {
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

func (c *CommissionerProposal) DynamicallyValid(ctx context.Context, db database.Provider) error {
	// this will return an error if the SeasonId on this proposal is not valid,
	// so we don't need to check for that scenario again in this function
	possibleVoters, err := c.GetAllVoterIds(ctx, db)
	if err != nil {
		return err
	}

	// validate that each recorded voter is a valid captain or commissioner
	rows, err := c.getVoteRows(ctx, db)
	if err != nil {
		return err
	}
	for _, row := range rows {
		found := false
		for _, possible := range possibleVoters {
			if row.UserId == possible {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("voter %s not found in possible voters for commissioner proposal", row.UserId)
		}
	}
	return nil
}

func (c *CommissionerProposal) GetAllVoterIds(ctx context.Context, db database.Provider) ([]database.UserId, error) {
	// get underlying season
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, c.SeasonId.RecordId())
	if err != nil {
		return nil, err
	}
	// add all commissioners as valid voters
	output := make([]database.UserId, 0)
	commissioners, err := season.GetCommissioners(ctx, db)
	if err == nil {
		output = append(output, commissioners...)
	}

	// get all team captains as the other voters
	teamCaptains, err := season.GetTeamCaptains(ctx, db)
	if err != nil {
		return nil, err
	}
	return append(output, teamCaptains...), nil
}

// getVoteRows returns all CommissionerProposalVote relationship rows assigned to
// this proposal.
func (c *CommissionerProposal) getVoteRows(ctx context.Context, db database.Provider) ([]*CommissionerProposalVote, error) {
	filter := func(_ context.Context, v *CommissionerProposalVote) bool {
		return v.ProposalId == c.ID
	}
	return database.GetAllWhere[*CommissionerProposalVote](ctx, db, filter)
}

// GetVotes reassembles the relationship rows into the former inline map shape
// (user id -> vote) for in-process reads.
func (c *CommissionerProposal) GetVotes(ctx context.Context, db database.Provider) (map[database.UserId]bool, error) {
	rows, err := c.getVoteRows(ctx, db)
	if err != nil {
		return nil, err
	}
	result := make(map[database.UserId]bool, len(rows))
	for _, row := range rows {
		result[row.UserId] = row.Vote
	}
	return result, nil
}

func (c *CommissionerProposal) Vote(ctx context.Context, db database.Provider, voterId database.UserId, vote bool) error {
	// only an entitled voter (a commissioner or team captain in the Season) may
	// cast a vote on this proposal
	possibleVoters, err := c.GetAllVoterIds(ctx, db)
	if err != nil {
		return err
	}
	valid := false
	for _, possible := range possibleVoters {
		if voterId == possible {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("voter %s not found in possible voters for commissioner proposal", voterId)
	}

	// if this voter has already voted, update their existing relationship row
	// rather than creating a duplicate (one vote per voter, enforced by the
	// natural unique constraint on (ProposalId, UserId))
	rows, err := c.getVoteRows(ctx, db)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.UserId == voterId {
			row.Vote = vote
			return database.UpdateOne(ctx, db, row)
		}
	}

	row := &CommissionerProposalVote{
		ProposalId: c.ID,
		UserId:     voterId,
		Vote:       vote,
	}
	_, err = database.CreateOne(ctx, db, row)
	return err
}

func (c *CommissionerProposal) VotesToPassOrFail(voterIds []database.UserId) (votesToPass, votesToFail int) {
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

func (c *CommissionerProposal) Status(ctx context.Context, db database.Provider) (accepted, rejected bool, err error) {
	voterIds, err := c.GetAllVoterIds(ctx, db)
	if err != nil {
		return false, false, err
	}

	// X votes to pass, Y votes to fail based on the proposal type
	votesNeededToPass, votesNeededToFail := c.VotesToPassOrFail(voterIds)

	rows, err := c.getVoteRows(ctx, db)
	if err != nil {
		return false, false, err
	}

	votesInFavor := 0
	votesAgainst := 0
	for _, row := range rows {
		// tally all the existing votes
		if row.Vote {
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

func (c *CommissionerProposal) NewRecord() database.CrudRecord {
	return new(CommissionerProposal)
}

// CommissionerProposalVote is a relationship record that assigns a single vote
// (for or against) from a User on a CommissionerProposal. This replaces the
// former denormalized `CommissionerProposal.Votes` inline map, enabling each
// vote to be queried/indexed individually and stored in its own collection.
//
// It is stored in its own collection (`commissioner_proposal_vote`) with FKs to
// commissioner_proposal and user, and a natural unique constraint on
// (ProposalId, UserId) guaranteeing one vote per voter. When the #36 SQLite
// provider lands, this becomes a `commissioner_proposal_votes` table with a
// migration/backfill.
type CommissionerProposalVote struct {
	ID         database.RecordId
	ProposalId database.RecordId
	UserId     database.UserId
	Vote       bool
}

func (v *CommissionerProposalVote) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (v *CommissionerProposalVote) SetOwner(userId database.UserId) {}

func (v *CommissionerProposalVote) Type() string {
	return "commissioner_proposal_vote"
}

func (v *CommissionerProposalVote) GetId() database.RecordId {
	return v.ID
}

func (v *CommissionerProposalVote) SetId(id database.RecordId) {
	v.ID = id
}

func (v *CommissionerProposalVote) StaticallyValid() error {
	return nil
}

func (v *CommissionerProposalVote) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &CommissionerProposal{}, v.ProposalId); err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &User{}, v.UserId.RecordId())
}

// AccessibleTo exposes the vote only to those who can view the proposal itself
// (the Season's commissioners and team captains).
func (v *CommissionerProposalVote) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	proposal, err := database.GetExistingRecordById(ctx, db, &CommissionerProposal{}, v.ProposalId)
	if err != nil {
		return nil
	}
	return proposal.AccessibleTo(ctx, db)
}

func (v *CommissionerProposalVote) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (v *CommissionerProposalVote) NewRecord() database.CrudRecord {
	return new(CommissionerProposalVote)
}

// UniquenessEquivalent enforces the natural unique constraint on
// (ProposalId, UserId): a user may only cast one vote per proposal.
func (v *CommissionerProposalVote) UniquenessEquivalent(other *CommissionerProposalVote) error {
	if v.ProposalId == other.ProposalId && v.UserId == other.UserId {
		return fmt.Errorf("voter %s already has a vote on commissioner proposal %s", v.UserId, v.ProposalId)
	}
	return nil
}
