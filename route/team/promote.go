package team

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for Team records. It matches the model's
// Type() ("team").
var BaseRoute = "/team"

// PromoteCoCaptainBody is the request body for PromoteCoCaptain.
type PromoteCoCaptainBody struct {
	// UserId is the team member to promote to co-captain.
	UserId database.UserId `json:"user_id"`
}

// StaticallyValid ensures a user was provided.
func (b *PromoteCoCaptainBody) StaticallyValid() error {
	if b.UserId == database.InvalidUserId {
		return errors.New("user_id must not be empty")
	}
	return nil
}

// PromoteCoCaptain promotes an existing team member to co-captain by updating
// their TeamAssignment role. This is a constrained role-assignment endpoint:
// it is the only way to change a team's roster roles (there is no generic
// create/update/delete on Team / TeamAssignment records), and only the team's
// captain / co-captains may use it.
//
// The captain cannot be demoted and members who are already co-captains are
// rejected, keeping the roster largely immutable after the draft finalizes.
type PromoteCoCaptain struct{}

func (c PromoteCoCaptain) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/promote_co_captain"
}

func (c PromoteCoCaptain) RequestBody() (*PromoteCoCaptainBody, bool) {
	return &PromoteCoCaptainBody{}, true
}

func (c PromoteCoCaptain) Handler(req api.Request[*PromoteCoCaptainBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	team, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Team{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Only the team's captain / co-captains may assign roles.
	wac := database.NewWithAccessControl[*model.Team](req.Context, req.DatabaseProvider, req.Token.UserId)
	if !wac.CanUserEdit(team) {
		return nil, http.StatusForbidden, errors.New("not authorized to modify this team")
	}

	assignment, err := findAssignment(req, team.ID, req.Body.UserId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if assignment == nil {
		return nil, http.StatusNotFound, errors.New("user is not a member of this team")
	}

	if assignment.Role == model.TeamRoleCaptain {
		return nil, http.StatusBadRequest, errors.New("the captain cannot be promoted to co-captain")
	}
	if assignment.Role == model.TeamRoleCoCaptain {
		return nil, http.StatusBadRequest, errors.New("user is already a co-captain of this team")
	}

	assignment.Role = model.TeamRoleCoCaptain
	if err := database.UpdateOne(req.Context, req.DatabaseProvider, assignment); err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: assignment}, http.StatusOK, nil
}

// findAssignment looks up the TeamAssignment join record linking the given
// team and user, returning nil if it does not exist.
func findAssignment[T database.Validatable](req api.Request[T], teamId model.TeamId, userId database.UserId) (*model.TeamAssignment, error) {
	assignments, err := database.GetAllWhere[*model.TeamAssignment](req.Context, req.DatabaseProvider,
		func(_ context.Context, a *model.TeamAssignment) bool {
			return a.TeamId == teamId && a.UserId == userId
		})
	if err != nil {
		return nil, fmt.Errorf("failed to look up team assignment: %w", err)
	}
	if len(assignments) == 0 {
		return nil, nil
	}
	return assignments[0], nil
}
