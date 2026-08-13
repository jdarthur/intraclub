package availability

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base path for the Availability REST surface. It matches the
// singular route convention used by the generic CRUD routes (derived from
// Availability.Type()).
const BaseRoute = "/availability"

// EmptyBody is used by routes that do not accept a request body.
type EmptyBody struct{}

func (b *EmptyBody) StaticallyValid() error {
	return nil
}

// SetAvailabilityBody is the request body for SetAvailability.
type SetAvailabilityBody struct {
	// WeekId is the week the availability applies to.
	WeekId model.WeekId `json:"week_id"`
	// Available is the participant's availability option for the week.
	Available model.AvailabilityOption `json:"available"`
}

// StaticallyValid ensures the availability option is a valid one before any
// handler logic runs.
func (b *SetAvailabilityBody) StaticallyValid() error {
	if !b.Available.Valid() {
		return errors.New("available must be a valid availability option")
	}
	return nil
}

// isSeasonParticipant reports whether the given user participates in the season
// the week belongs to (a commissioner, late addition, or member of a season
// team). Only participants may set availability.
func isSeasonParticipant(ctx context.Context, db database.Provider, week *model.Week, userId database.UserId) (bool, error) {
	draft, err := database.GetExistingRecordById(ctx, db, &model.Draft{}, week.DraftId.RecordId())
	if err != nil {
		return false, err
	}
	season, err := draft.GetSeason(ctx, db)
	if err != nil {
		return false, err
	}
	if season == nil {
		// No season exists yet for the week's draft, so there are no
		// participants to set availability for.
		return false, nil
	}
	return season.IsUserIdASeasonParticipant(ctx, db, userId)
}

// SetAvailability creates or updates the requesting user's availability for a
// week. Only the user themselves may set their availability (the record owner
// is always the requesting user), and only if they are a participant in the
// week's season. If the user has already set availability for the week, the
// existing record is updated (upsert); otherwise a new record is created. The
// per-user+week uniqueness rule in the model prevents duplicates.
type SetAvailability struct{}

func (c SetAvailability) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute + "/set"
}

func (c SetAvailability) RequestBody() (*SetAvailabilityBody, bool) {
	return &SetAvailabilityBody{}, true
}

func (c SetAvailability) Handler(req api.Request[*SetAvailabilityBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	week, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Week{}, req.Body.WeekId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	participant, err := isSeasonParticipant(req.Context, req.DatabaseProvider, week, req.Token.UserId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if !participant {
		return nil, http.StatusForbidden, errors.New("only a season participant may set availability")
	}

	// Upsert: update the existing record for this user+week if present,
	// otherwise create a new one. Uniqueness (user+week) is enforced by the
	// model's UniquenessEquivalent, which CreateOne checks against existing
	// records.
	existing, err := database.GetAllWhere[*model.Availability](req.Context, req.DatabaseProvider,
		func(_ context.Context, a *model.Availability) bool {
			return a.UserId == req.Token.UserId && a.WeekId == req.Body.WeekId
		})
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if len(existing) > 0 {
		existing[0].Available = req.Body.Available
		if err := database.UpdateOne(req.Context, req.DatabaseProvider, existing[0]); err != nil {
			return nil, http.StatusBadRequest, err
		}
		return gin.H{api.ResourceKey: existing[0]}, http.StatusOK, nil
	}

	availability := model.NewAvailability()
	availability.UserId = req.Token.UserId
	availability.WeekId = req.Body.WeekId
	availability.Available = req.Body.Available
	created, err := database.CreateOne(req.Context, req.DatabaseProvider, availability)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: created}, http.StatusOK, nil
}

// GetForUser returns the requesting user's (or a specified user's) availability
// for the weeks of a draft. If user_id is omitted, the requesting user's own
// availability is returned. Only the user themselves or a system administrator
// may view a user's availability.
type GetForUser struct{}

func (c GetForUser) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, BaseRoute
}

func (c GetForUser) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c GetForUser) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	query := req.HTTPRequest().URL.Query()
	draftStr := query.Get("draft_id")
	if draftStr == "" {
		return nil, http.StatusBadRequest, errors.New("draft_id query parameter is required")
	}
	draftId, err := database.RecordIdFromString(draftStr)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	userId := req.Token.UserId
	if uidStr := query.Get("user_id"); uidStr != "" {
		parsed, err := database.RecordIdFromString(uidStr)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		if database.UserId(parsed) != req.Token.UserId {
			isSysAdmin, err := database.SysAdminCheck(req.Context, req.DatabaseProvider, req.Token.UserId)
			if err != nil {
				return nil, http.StatusBadRequest, err
			}
			if !isSysAdmin {
				return nil, http.StatusForbidden, errors.New("only a user may view their own availability")
			}
		}
		userId = database.UserId(parsed)
	}

	avail, err := model.GetAvailabilityForUser(req.Context, req.DatabaseProvider, userId, model.DraftId(draftId))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: avail}, http.StatusOK, nil
}

// isCaptainOrCoCaptain reports whether the given user is the team's captain or
// a co-captain, or a system administrator. Team availability is only viewable
// by these roles.
func isCaptainOrCoCaptain(ctx context.Context, db database.Provider, team *model.Team, userId database.UserId) (bool, error) {
	captain, err := team.GetCaptain(ctx, db)
	if err != nil {
		return false, err
	}
	if captain == userId {
		return true, nil
	}

	coCaptains, err := team.GetCoCaptains(ctx, db)
	if err != nil {
		return false, err
	}
	for _, c := range coCaptains {
		if c == userId {
			return true, nil
		}
	}

	isSysAdmin, err := database.SysAdminCheck(ctx, db, userId)
	if err != nil {
		return false, err
	}
	return isSysAdmin, nil
}

// TeamAvailabilityEntry is a single user's availability within the team
// availability response. It is used instead of a map keyed by user ID because
// Go's encoding/json serializes map keys of integer-backed types as decimal,
// which would not match the hex user IDs used everywhere else in the API.
type TeamAvailabilityEntry struct {
	UserId         database.UserId       `json:"user_id"`
	Availabilities []*model.Availability `json:"availabilities"`
}

// GetForTeam returns per-user availability for every member of a team across
// the weeks of a draft. Only the team's captain / co-captains (or a system
// administrator) may view it. The response is a list of { user_id,
// availabilities } entries, one per team member.
type GetForTeam struct{}

func (c GetForTeam) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, BaseRoute + "/team"
}

func (c GetForTeam) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c GetForTeam) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	query := req.HTTPRequest().URL.Query()
	teamStr := query.Get("team_id")
	if teamStr == "" {
		return nil, http.StatusBadRequest, errors.New("team_id query parameter is required")
	}
	teamId, err := database.RecordIdFromString(teamStr)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	draftStr := query.Get("draft_id")
	if draftStr == "" {
		return nil, http.StatusBadRequest, errors.New("draft_id query parameter is required")
	}
	draftId, err := database.RecordIdFromString(draftStr)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	team, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Team{}, teamId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	authorized, err := isCaptainOrCoCaptain(req.Context, req.DatabaseProvider, team, req.Token.UserId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if !authorized {
		return nil, http.StatusForbidden, errors.New("only a team captain or co-captain may view team availability")
	}

	avail, err := model.GetAvailabilityForTeam(req.Context, req.DatabaseProvider, model.TeamId(teamId), model.DraftId(draftId))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	entries := make([]*TeamAvailabilityEntry, 0, len(avail))
	for userId, list := range avail {
		entries = append(entries, &TeamAvailabilityEntry{UserId: userId, Availabilities: list})
	}
	return gin.H{api.ResourceKey: entries}, http.StatusOK, nil
}
