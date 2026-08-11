package draft

import (
	"errors"
	"net/http"
	"time"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for Draft records.
var BaseRoute = "/draft"

// loadEditableDraft fetches the Draft referenced by the request's path ID and
// verifies the requesting user (from the token) is authorized to edit it
// (owner / sysadmin). This mirrors the access-control pattern used by the
// Format and Ruleset custom routes.
func loadEditableDraft[T database.Validatable](req api.Request[T]) (*model.Draft, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	draft, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Draft{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	wac := database.NewWithAccessControl[*model.Draft](req.Context, req.DatabaseProvider, req.Token.UserId)
	if !wac.CanUserEdit(draft) {
		return nil, http.StatusForbidden, errors.New("not authorized to modify this draft")
	}

	return draft, http.StatusOK, nil
}

// EmptyBody is used by routes that do not accept a request body.
type EmptyBody struct{}

func (b *EmptyBody) StaticallyValid() error {
	return nil
}

// InitializeBody is the request body for InitializeDraft.
type InitializeBody struct {
	// Captains are the UserIds who will captain the draft's teams, in draft
	// order.
	Captains []database.UserId `json:"captains"`
}

// StaticallyValid ensures at least one captain was provided.
func (b *InitializeBody) StaticallyValid() error {
	if len(b.Captains) == 0 {
		return errors.New("captains must not be empty")
	}
	return nil
}

// InitializeDraft creates the draft's teams, DraftCaptain assignments and
// DraftAvailablePlayer rows from the provided captain list (see
// model.Draft.Initialize).
type InitializeDraft struct{}

func (c InitializeDraft) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/initialize"
}

func (c InitializeDraft) RequestBody() (*InitializeBody, bool) {
	return &InitializeBody{}, true
}

func (c InitializeDraft) Handler(req api.Request[*InitializeBody]) (any, int, error) {
	draft, status, err := loadEditableDraft(req)
	if err != nil {
		return nil, status, err
	}

	if err := draft.Initialize(req.Context, req.DatabaseProvider, req.Body.Captains); err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: draft}, http.StatusOK, nil
}

// AssignDraftablePlayersBody is the request body for AssignDraftablePlayers.
type AssignDraftablePlayersBody struct {
	// Players are the UserIds to add to the draft's available-to-draft list.
	Players []database.UserId `json:"players"`
}

// StaticallyValid ensures at least one player was provided.
func (b *AssignDraftablePlayersBody) StaticallyValid() error {
	if len(b.Players) == 0 {
		return errors.New("players must not be empty")
	}
	return nil
}

// AssignDraftablePlayers adds players to the draft's available-to-draft list
// (see model.Draft.AssignDraftablePlayers).
type AssignDraftablePlayers struct{}

func (c AssignDraftablePlayers) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/assign_draftable_players"
}

func (c AssignDraftablePlayers) RequestBody() (*AssignDraftablePlayersBody, bool) {
	return &AssignDraftablePlayersBody{}, true
}

func (c AssignDraftablePlayers) Handler(req api.Request[*AssignDraftablePlayersBody]) (any, int, error) {
	draft, status, err := loadEditableDraft(req)
	if err != nil {
		return nil, status, err
	}

	if err := draft.AssignDraftablePlayers(req.Context, req.DatabaseProvider, req.Body.Players); err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: draft}, http.StatusOK, nil
}

// AssignRatingCutoffBody is the request body for AssignRatingCutoff.
type AssignRatingCutoffBody struct {
	// Rating is the RatingId to assign a cutoff index to.
	Rating database.RecordId `json:"rating"`
	// Cutoff is the last selection index matching this rating.
	Cutoff int `json:"cutoff"`
}

// StaticallyValid ensures the cutoff is a positive index.
func (b *AssignRatingCutoffBody) StaticallyValid() error {
	if b.Cutoff <= 0 {
		return errors.New("cutoff must be greater than zero")
	}
	return nil
}

// AssignRatingCutoff creates the DraftRatingCutoff row assigning a cutoff index
// to a rating for the draft (see model.Draft.AssignRatingCutoff).
type AssignRatingCutoff struct{}

func (c AssignRatingCutoff) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/assign_rating_cutoff"
}

func (c AssignRatingCutoff) RequestBody() (*AssignRatingCutoffBody, bool) {
	return &AssignRatingCutoffBody{}, true
}

func (c AssignRatingCutoff) Handler(req api.Request[*AssignRatingCutoffBody]) (any, int, error) {
	draft, status, err := loadEditableDraft(req)
	if err != nil {
		return nil, status, err
	}

	row, err := draft.AssignRatingCutoff(req.Context, req.DatabaseProvider, model.RatingId(req.Body.Rating), req.Body.Cutoff)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: row}, http.StatusOK, nil
}

// SelectBody is the request body for SelectByCaptain.
type SelectBody struct {
	// PlayerId is the UserId being selected by the captain on the clock.
	PlayerId database.UserId `json:"player_id"`
}

// StaticallyValid ensures a player was provided.
func (b *SelectBody) StaticallyValid() error {
	if b.PlayerId == database.InvalidUserId {
		return errors.New("player_id must not be empty")
	}
	return nil
}

// SelectByCaptain makes a draft pick on behalf of the authenticated captain
// (see model.Draft.SelectByCaptain). The requesting user must be a captain
// assigned to this draft and the captain currently on the clock.
type SelectByCaptain struct{}

func (c SelectByCaptain) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/select"
}

func (c SelectByCaptain) RequestBody() (*SelectBody, bool) {
	return &SelectBody{}, true
}

func (c SelectByCaptain) Handler(req api.Request[*SelectBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	draft, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Draft{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// The requesting user must be a captain assigned to this draft.
	captains, err := draft.GetCaptains(req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	isCaptain := false
	for _, c := range captains {
		if c.CaptainId == req.Token.UserId {
			isCaptain = true
			break
		}
	}
	if !isCaptain {
		return nil, http.StatusForbidden, errors.New("only a draft captain may make a selection")
	}

	if err := draft.SelectByCaptain(req.Context, req.Body.PlayerId, req.Token.UserId, req.DatabaseProvider); err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: draft}, http.StatusOK, nil
}

// AssignDraftedPlayersToTeams finalizes a completed draft by assigning each
// drafted player to their team with their draft rating (see
// model.Draft.AssignDraftedPlayersToTeams).
type AssignDraftedPlayersToTeams struct{}

func (c AssignDraftedPlayersToTeams) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/assign_drafted_players_to_teams"
}

func (c AssignDraftedPlayersToTeams) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c AssignDraftedPlayersToTeams) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	draft, status, err := loadEditableDraft(req)
	if err != nil {
		return nil, status, err
	}

	if err := draft.AssignDraftedPlayersToTeams(req.Context, req.DatabaseProvider); err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: draft}, http.StatusOK, nil
}

// CreateSeasonBody is the request body for CreateSeason.
type CreateSeasonBody struct {
	// Name is the name given to the new Season.
	Name string `json:"name"`
	// Facility is the FacilityId the Season is played at.
	Facility database.RecordId `json:"facility"`
	// StartTime is the daily kickoff time in 24-hour "HH:MM" format (e.g. "08:30").
	StartTime string `json:"start_time"`
}

// StaticallyValid ensures the season's name is present.
func (b *CreateSeasonBody) StaticallyValid() error {
	if b.Name == "" {
		return errors.New("name must not be empty")
	}
	return nil
}

// CreateSeason creates the Season associated with a completed draft (see
// model.Draft.CreateSeason).
type CreateSeason struct{}

func (c CreateSeason) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/create_season"
}

func (c CreateSeason) RequestBody() (*CreateSeasonBody, bool) {
	return &CreateSeasonBody{}, true
}

func (c CreateSeason) Handler(req api.Request[*CreateSeasonBody]) (any, int, error) {
	draft, status, err := loadEditableDraft(req)
	if err != nil {
		return nil, status, err
	}

	hour, minute, err := parseStartTime(req.Body.StartTime)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	season, err := draft.CreateSeason(req.Context, req.DatabaseProvider, req.Body.Name, model.FacilityId(req.Body.Facility), model.NewStartTime(hour, minute))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: season}, http.StatusOK, nil
}

// SetDraftOrderPatternBody is the request body for SetDraftOrderPattern. The
// Draft.DraftOrderPattern field is an interface and cannot be deserialized
// directly, so it is selected by its Name() string (e.g. "Snake").
type SetDraftOrderPatternBody struct {
	// DraftOrderPattern is the Name() of one of the draft order patterns
	// returned by GET /api/draft_order_patterns.
	DraftOrderPattern string `json:"draft_order_pattern"`
}

// StaticallyValid ensures a pattern name was provided.
func (b *SetDraftOrderPatternBody) StaticallyValid() error {
	if b.DraftOrderPattern == "" {
		return errors.New("draft_order_pattern must not be empty")
	}
	return nil
}

// SetDraftOrderPattern sets the draft's DraftOrderPattern by its Name() string.
type SetDraftOrderPattern struct{}

func (c SetDraftOrderPattern) Path() (api.HttpMethod, string) {
	return api.HttpMethodPut, api.AppendPathId(BaseRoute) + "/draft_order_pattern"
}

func (c SetDraftOrderPattern) RequestBody() (*SetDraftOrderPatternBody, bool) {
	return &SetDraftOrderPatternBody{}, true
}

func (c SetDraftOrderPattern) Handler(req api.Request[*SetDraftOrderPatternBody]) (any, int, error) {
	draft, status, err := loadEditableDraft(req)
	if err != nil {
		return nil, status, err
	}

	pattern, err := model.DraftOrderPatternFromString(req.Body.DraftOrderPattern)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	draft.DraftOrderPattern = pattern
	if err := database.UpdateOne(req.Context, req.DatabaseProvider, draft); err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: draft}, http.StatusOK, nil
}

// parseStartTime parses a 24-hour "HH:MM" string into its hour and minute
// components.
func parseStartTime(raw string) (hour int, minute int, err error) {
	if raw == "" {
		return 0, 0, errors.New("start_time must not be empty")
	}
	t, err := time.Parse("15:04", raw)
	if err != nil {
		return 0, 0, errors.New("start_time must be in 24-hour HH:MM format (e.g. 08:30)")
	}
	return t.Hour(), t.Minute(), nil
}
