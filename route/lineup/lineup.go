package lineup

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base path for the Lineup REST surface. It matches the
// singular route convention used by the generic CRUD routes (derived from
// Lineup.Type()).
const BaseRoute = "/lineup"

// canEditTeamLineup reports whether the requesting user is a captain or
// co-captain of the given team (the only role allowed to build/confirm it).
func canEditTeamLineup(ctx context.Context, db database.Provider, userId database.UserId, teamId model.TeamId) bool {
	for _, uid := range model.EditableByTeamCaptainOrCoCaptains(ctx, db, teamId) {
		if uid == userId {
			return true
		}
	}
	return false
}

// isSeasonCommissioner reports whether the requesting user is one of the
// season's commissioners (the only role allowed to mark a lineup official).
func isSeasonCommissioner(ctx context.Context, db database.Provider, userId database.UserId, seasonId model.SeasonId) bool {
	for _, uid := range model.EditableBySeason(ctx, db, seasonId) {
		if uid == userId {
			return true
		}
	}
	return false
}

// lineupDetailForTeamWeek builds the LineupDetail for a team + week: the
// lineup record (if any), the format's lines (in index order), and the
// lineup's pairings.
func lineupDetailForTeamWeek(ctx context.Context, db database.Provider, teamId model.TeamId, weekId model.WeekId) (*LineupDetail, error) {
	week, err := database.GetExistingRecordById(ctx, db, &model.Week{}, weekId.RecordId())
	if err != nil {
		return nil, err
	}
	draft, err := database.GetExistingRecordById(ctx, db, &model.Draft{}, week.DraftId.RecordId())
	if err != nil {
		return nil, err
	}
	format, err := database.GetExistingRecordById(ctx, db, &model.Format{}, draft.Format.RecordId())
	if err != nil {
		return nil, err
	}
	lines, err := format.GetLines(ctx, db)
	if err != nil {
		return nil, err
	}

	detail := &LineupDetail{
		TeamId:   teamId,
		WeekId:   weekId,
		Lines:    lines,
		Pairings: []*model.LineupPairing{},
	}

	existing, err := database.GetAllWhere[*model.Lineup](ctx, db, func(_ context.Context, l *model.Lineup) bool {
		return l.TeamId == teamId && l.WeekId == weekId
	})
	if err != nil {
		return nil, err
	}
	if len(existing) == 0 {
		return detail, nil
	}
	lineup := existing[0]
	detail.Lineup = lineup

	pairings, err := database.GetAllWhere[*model.LineupPairing](ctx, db, func(_ context.Context, p *model.LineupPairing) bool {
		return p.LineupId == lineup.ID
	})
	if err != nil {
		return nil, err
	}
	detail.Pairings = pairings
	return detail, nil
}

// LineupDetail is the wire representation of a team's weekly lineup, including
// the format lines the captain must fill and the pairings already assigned.
type LineupDetail struct {
	TeamId   model.TeamId            `json:"team_id"`
	WeekId   model.WeekId            `json:"week_id"`
	Lineup   *model.Lineup           `json:"lineup"`
	Lines    []model.FormatLine      `json:"lines"`
	Pairings []*model.LineupPairing  `json:"pairings"`
}

// LineupQuery holds the query parameters for GetLineupDetail.
type LineupQuery struct {
	TeamId model.TeamId `json:"team_id"`
	WeekId model.WeekId `json:"week_id"`
}

// StaticallyValid ensures both a team and a week are specified.
func (q *LineupQuery) StaticallyValid() error {
	if q.TeamId.RecordId() == database.InvalidRecordId {
		return errors.New("team_id must be set")
	}
	if q.WeekId.RecordId() == database.InvalidRecordId {
		return errors.New("week_id must be set")
	}
	return nil
}

// GetLineupDetail returns the LineupDetail for a team + week. It is viewable by
// everyone (team members and commissioners alike).
type GetLineupDetail struct{}

func (c GetLineupDetail) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, BaseRoute + "/detail"
}

func (c GetLineupDetail) RequestBody() (*LineupQuery, bool) {
	return &LineupQuery{}, false
}

func (c GetLineupDetail) Handler(req api.Request[*LineupQuery]) (any, int, error) {
	query := req.HTTPRequest().URL.Query()

	var teamId model.TeamId
	var weekId model.WeekId

	if teamStr := query.Get("team_id"); teamStr != "" {
		rid, err := database.RecordIdFromString(teamStr)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		teamId = model.TeamId(rid)
	}
	if weekStr := query.Get("week_id"); weekStr != "" {
		rid, err := database.RecordIdFromString(weekStr)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		weekId = model.WeekId(rid)
	}

	if teamId.RecordId() == database.InvalidRecordId {
		return nil, http.StatusBadRequest, errors.New("team_id must be set")
	}
	if weekId.RecordId() == database.InvalidRecordId {
		return nil, http.StatusBadRequest, errors.New("week_id must be set")
	}

	detail, err := lineupDetailForTeamWeek(req.Context, req.DatabaseProvider, teamId, weekId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: detail}, http.StatusOK, nil
}

// PairingInput is a single format-line assignment in a SetLineup request.
type PairingInput struct {
	Player1         database.UserId `json:"player1"`
	Player2         database.UserId `json:"player2"`
	FormatLineIndex int             `json:"format_line_index"`
}

// SetLineupBody is the request body for SetLineup.
type SetLineupBody struct {
	TeamId   model.TeamId     `json:"team_id"`
	WeekId   model.WeekId     `json:"week_id"`
	Pairings []PairingInput   `json:"pairings"`
}

// StaticallyValid ensures both a team and a week are specified.
func (b *SetLineupBody) StaticallyValid() error {
	if b.TeamId.RecordId() == database.InvalidRecordId {
		return errors.New("team_id must be set")
	}
	if b.WeekId.RecordId() == database.InvalidRecordId {
		return errors.New("week_id must be set")
	}
	return nil
}

// SetLineup creates (or updates) a team's weekly lineup and replaces its
// pairings. Only the team's captain/co-captains may build a lineup. Each
// pairing is validated against team membership and format ratings by the model.
type SetLineup struct{}

func (c SetLineup) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute + "/set"
}

func (c SetLineup) RequestBody() (*SetLineupBody, bool) {
	return &SetLineupBody{}, true
}

func (c SetLineup) Handler(req api.Request[*SetLineupBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}
	if !canEditTeamLineup(req.Context, req.DatabaseProvider, req.Token.UserId, req.Body.TeamId) {
		return nil, http.StatusForbidden, errors.New("only a team captain or co-captain may build the lineup")
	}

	// Find (or create) the Lineup for this team + week.
	var lineup *model.Lineup
	existing, err := database.GetAllWhere[*model.Lineup](req.Context, req.DatabaseProvider, func(_ context.Context, l *model.Lineup) bool {
		return l.TeamId == req.Body.TeamId && l.WeekId == req.Body.WeekId
	})
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if len(existing) > 0 {
		lineup = existing[0]
	} else {
		lineup = model.NewLineup()
		lineup.TeamId = req.Body.TeamId
		lineup.WeekId = req.Body.WeekId
		lineup, err = database.CreateOne(req.Context, req.DatabaseProvider, lineup)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
	}

	// Replace the lineup's pairings. Deleting the existing rows first means an
	// empty set clears the lineup; creating fresh rows re-validates each one.
	oldPairings, err := database.GetAllWhere[*model.LineupPairing](req.Context, req.DatabaseProvider, func(_ context.Context, p *model.LineupPairing) bool {
		return p.LineupId == lineup.ID
	})
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	for _, p := range oldPairings {
		if _, _, err := database.DeleteOneById(req.Context, req.DatabaseProvider, p, p.ID.RecordId()); err != nil {
			return nil, http.StatusBadRequest, err
		}
	}

	for _, in := range req.Body.Pairings {
		pairing := model.NewLineupPairing()
		pairing.LineupId = lineup.ID
		pairing.TeamId = req.Body.TeamId
		pairing.Player1 = in.Player1
		pairing.Player2 = in.Player2
		pairing.FormatLineIndex = in.FormatLineIndex
		if _, err := database.CreateOne(req.Context, req.DatabaseProvider, pairing); err != nil {
			return nil, http.StatusBadRequest, err
		}
	}

	detail, err := lineupDetailForTeamWeek(req.Context, req.DatabaseProvider, req.Body.TeamId, req.Body.WeekId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: detail}, http.StatusOK, nil
}

// ConfirmLineup marks a lineup as confirmed by the team's captain/co-captains,
// which is required before the commissioner can mark it official.
type ConfirmLineup struct{}

func (c ConfirmLineup) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/confirm"
}

func (c ConfirmLineup) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c ConfirmLineup) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}
	lineup, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Lineup{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if !canEditTeamLineup(req.Context, req.DatabaseProvider, req.Token.UserId, lineup.TeamId) {
		return nil, http.StatusForbidden, errors.New("only a team captain or co-captain may confirm the lineup")
	}
	lineup.Confirmed = true
	if err := database.UpdateOne(req.Context, req.DatabaseProvider, lineup); err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: lineup}, http.StatusOK, nil
}

// MarkOfficial marks a confirmed lineup as official. Only a season
// commissioner may do this, and the lineup must first be confirmed.
type MarkOfficial struct{}

func (c MarkOfficial) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/official"
}

func (c MarkOfficial) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c MarkOfficial) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}
	lineup, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Lineup{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if !lineup.Confirmed {
		return nil, http.StatusBadRequest, errors.New("lineup must be confirmed before it can be marked official")
	}

	week, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Week{}, lineup.WeekId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	draft, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Draft{}, week.DraftId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	season, err := draft.GetSeason(req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if !isSeasonCommissioner(req.Context, req.DatabaseProvider, req.Token.UserId, season.ID) {
		return nil, http.StatusForbidden, errors.New("only a season commissioner may mark the lineup official")
	}

	lineup.Official = true
	if err := database.UpdateOne(req.Context, req.DatabaseProvider, lineup); err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: lineup}, http.StatusOK, nil
}

// EmptyBody is a placeholder request body for routes that take no body.
type EmptyBody struct{}

// StaticallyValid has no static constraints.
func (b *EmptyBody) StaticallyValid() error { return nil }
