// Package proposal implements the custom REST surface for
// model.CommissionerProposal: casting a vote (validating the voter is a season
// participant) and fetching a proposal's detail/tally. The generic CRUD create
// / list / update / delete endpoints are registered in main.go.
package proposal

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for CommissionerProposal records. It matches
// the model's Type() ("commissioner_proposal").
var BaseRoute = "/commissioner_proposal"

// ProposalDetail is the read shape returned for a single proposal: the record
// itself plus its vote tally, the set of eligible voters, the pass/fail status,
// and (when the requester is eligible) their own vote.
type ProposalDetail struct {
	Proposal     *model.CommissionerProposal `json:"proposal"`
	VotesFor     int                         `json:"votes_for"`
	VotesAgainst int                         `json:"votes_against"`
	Voters       []database.UserId           `json:"voters"`
	Accepted     bool                        `json:"accepted"`
	Rejected     bool                        `json:"rejected"`
	// MyVote is the requesting user's vote, if they have cast one yet.
	MyVote *bool `json:"my_vote,omitempty"`
}

// RegisterRoutes wires up the custom proposal endpoints.
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	voteFamily := api.RouteFamily[*CastVoteBody]{DatabaseProvider: db}
	voteFamily.Handle(e, CastVote{})

	detailFamily := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	detailFamily.Handle(e, GetProposalDetail{})
}

// EmptyBody is used by the detail route, which accepts no request body.
type EmptyBody struct{}

func (b *EmptyBody) StaticallyValid() error {
	return nil
}

// CastVoteBody is the request body for CastVote.
type CastVoteBody struct {
	// Vote is true for "in favor", false for "against".
	Vote bool `json:"vote"`
}

func (b *CastVoteBody) StaticallyValid() error {
	return nil
}

// CastVote records the authenticated user's vote on a proposal. The model's
// Vote method enforces that the voter is a season participant (commissioner or
// team captain) and that each voter casts at most one vote (subsequent votes
// update the existing row).
type CastVote struct{}

func (c CastVote) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/vote"
}

func (c CastVote) RequestBody() (*CastVoteBody, bool) {
	return &CastVoteBody{}, true
}

func (c CastVote) Handler(req api.Request[*CastVoteBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	proposal, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.CommissionerProposal{}, req.PathId)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	if err := proposal.Vote(req.Context, req.DatabaseProvider, req.Token.UserId, req.Body.Vote); err != nil {
		return nil, http.StatusForbidden, err
	}

	detail, status, err := buildDetail(req.Context, req.DatabaseProvider, proposal, req.Token.UserId)
	if err != nil {
		return nil, status, err
	}
	return gin.H{api.ResourceKey: detail}, http.StatusOK, nil
}

// GetProposalDetail returns a proposal's detail and vote tally to any eligible
// voter (a commissioner or team captain of the proposal's season).
type GetProposalDetail struct{}

func (c GetProposalDetail) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, api.AppendPathId(BaseRoute) + "/detail"
}

func (c GetProposalDetail) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c GetProposalDetail) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	proposal, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.CommissionerProposal{}, req.PathId)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	// Only eligible voters may view the detail/tally.
	wac := database.NewWithAccessControl[*model.CommissionerProposal](req.Context, req.DatabaseProvider, req.Token.UserId)
	if !wac.CanUserAccess(proposal) {
		return nil, http.StatusNotFound, errors.New("proposal not found")
	}

	detail, status, err := buildDetail(req.Context, req.DatabaseProvider, proposal, req.Token.UserId)
	if err != nil {
		return nil, status, err
	}
	return gin.H{api.ResourceKey: detail}, http.StatusOK, nil
}

// buildDetail assembles the ProposalDetail response for a proposal.
func buildDetail(ctx context.Context, db database.Provider, proposal *model.CommissionerProposal, userId database.UserId) (*ProposalDetail, int, error) {
	voters, err := proposal.GetAllVoterIds(ctx, db)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	accepted, rejected, err := proposal.Status(ctx, db)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	votes, err := proposal.GetVotes(ctx, db)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	var myVote *bool
	if v, ok := votes[userId]; ok {
		vv := v
		myVote = &vv
	}

	votesFor := 0
	votesAgainst := 0
	for _, v := range votes {
		if v {
			votesFor++
		} else {
			votesAgainst++
		}
	}

	return &ProposalDetail{
		Proposal:     proposal,
		VotesFor:     votesFor,
		VotesAgainst: votesAgainst,
		Voters:       voters,
		Accepted:     accepted,
		Rejected:     rejected,
		MyVote:       myVote,
	}, http.StatusOK, nil
}
