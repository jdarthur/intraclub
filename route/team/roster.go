package team

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// TeamRoster is the read shape returned for a team: the Team record itself
// plus its member assignments (with roles). It is used by both the list and
// detail endpoints.
type TeamRoster struct {
	Team        *model.Team             `json:"team"`
	Assignments []*model.TeamAssignment `json:"assignments"`
}

// loadAccessibleTeam fetches the Team referenced by the request's path ID and
// verifies the requesting user may view it (member, sysadmin, or a season
// commissioner). Returns the team, or the appropriate HTTP status + error if
// the user cannot access it.
func loadAccessibleTeam[T database.Validatable](req api.Request[T]) (*model.Team, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	team, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Team{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	ok, err := canAccessTeam(req, team, req.Token.UserId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if !ok {
		return nil, http.StatusNotFound, errors.New("team not found")
	}
	return team, http.StatusOK, nil
}

// canAccessTeam reports whether the given user may view the team: its members,
// a sysadmin, or a commissioner of any season the team belongs to. Role
// assignment (promote) is a separate, stricter check in PromoteCoCaptain.
func canAccessTeam[T database.Validatable](req api.Request[T], team *model.Team, userId database.UserId) (bool, error) {
	isMember, err := team.IsTeamMember(req.Context, req.DatabaseProvider, userId)
	if err != nil {
		return false, err
	}
	if isMember {
		return true, nil
	}

	isSysAdmin, err := database.SysAdminCheck(req.Context, req.DatabaseProvider, userId)
	if err != nil {
		return false, err
	}
	if isSysAdmin {
		return true, nil
	}

	return isSeasonCommissionerOfTeam(req.Context, req.DatabaseProvider, team.ID, userId)
}

// isSeasonCommissionerOfTeam reports whether the user is a commissioner of any
// season that includes the given team.
func isSeasonCommissionerOfTeam(ctx context.Context, db database.Provider, teamId model.TeamId, userId database.UserId) (bool, error) {
	seasonTeams, err := database.GetAllWhere[*model.SeasonTeam](ctx, db,
		func(_ context.Context, st *model.SeasonTeam) bool { return st.TeamId == teamId })
	if err != nil {
		return false, err
	}
	if len(seasonTeams) == 0 {
		return false, nil
	}

	seasonIds := make(map[model.SeasonId]struct{}, len(seasonTeams))
	for _, st := range seasonTeams {
		seasonIds[st.SeasonId] = struct{}{}
	}

	commissioners, err := database.GetAllWhere[*model.SeasonCommissioner](ctx, db,
		func(_ context.Context, sc *model.SeasonCommissioner) bool { return sc.UserId == userId })
	if err != nil {
		return false, err
	}
	for _, sc := range commissioners {
		if _, ok := seasonIds[sc.SeasonId]; ok {
			return true, nil
		}
	}
	return false, nil
}

// buildRoster loads the team's member assignments (with roles).
func buildRoster[T database.Validatable](req api.Request[T], team *model.Team) ([]*model.TeamAssignment, error) {
	assignments, err := database.GetAllWhere[*model.TeamAssignment](req.Context, req.DatabaseProvider,
		func(_ context.Context, a *model.TeamAssignment) bool { return a.TeamId == team.ID })
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

// EmptyBody is used by routes that do not accept a request body.
type EmptyBody struct{}

func (b *EmptyBody) StaticallyValid() error {
	return nil
}

// ListTeams returns the teams the requesting user may view (members, sysadmins,
// and season commissioners), each with its roster. This is the read surface for
// the /teams list page. Team records are deliberately not exposed via generic
// CRUD.
type ListTeams struct{}

func (c ListTeams) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, BaseRoute
}

func (c ListTeams) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c ListTeams) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	teams, err := database.GetAll[*model.Team](req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	rosters := make([]TeamRoster, 0, len(teams))
	for _, team := range teams {
		ok, err := canAccessTeam(req, team, req.Token.UserId)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		if !ok {
			continue
		}
		assignments, err := buildRoster(req, team)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		rosters = append(rosters, TeamRoster{Team: team, Assignments: assignments})
	}

	return gin.H{api.ResourceKey: rosters}, http.StatusOK, nil
}

// GetTeam returns a single team and its roster, if the requesting user is a
// member (or sysadmin).
type GetTeam struct{}

func (c GetTeam) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, api.AppendPathId(BaseRoute)
}

func (c GetTeam) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c GetTeam) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	team, status, err := loadAccessibleTeam(req)
	if err != nil {
		return nil, status, err
	}
	assignments, err := buildRoster(req, team)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: TeamRoster{Team: team, Assignments: assignments}}, http.StatusOK, nil
}
