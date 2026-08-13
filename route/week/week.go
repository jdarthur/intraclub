package week

import (
	"errors"
	"net/http"
	"time"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base path for the Week REST surface. It matches the
// singular route convention used by the generic CRUD routes (derived from
// Week.Type()).
const BaseRoute = "/week"

// CreateWeekBody is the request body for CreateWeek.
type CreateWeekBody struct {
	// DraftId is the Draft the week belongs to.
	DraftId model.DraftId `json:"draft_id"`
	// Date is the scheduled playing date for the week.
	Date time.Time `json:"date"`
	// Note is an optional note for the week.
	Note string `json:"note"`
}

// StaticallyValid ensures the week's date is set before any handler logic runs.
func (b *CreateWeekBody) StaticallyValid() error {
	if b.Date.IsZero() {
		return errors.New("date must not be zero")
	}
	return nil
}

// CreateWeek creates a Week for a Draft. Weeks may only be created by a
// commissioner of the Season associated with the Draft (or, before a Season
// exists for the Draft, the Draft's owner) — normal participants and team
// captains are not authorized.
type CreateWeek struct{}

func (c CreateWeek) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute
}

func (c CreateWeek) RequestBody() (*CreateWeekBody, bool) {
	return &CreateWeekBody{}, true
}

func (c CreateWeek) Handler(req api.Request[*CreateWeekBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	week := model.NewWeek()
	week.DraftId = req.Body.DraftId
	week.Date = req.Body.Date
	week.Note = req.Body.Note

	// Resolve the users who are allowed to create/edit weeks for this Draft
	// (Season commissioners, or the Draft owner if no Season exists yet) and
	// require the requesting user to be among them.
	allowed := week.EditableBy(req.Context, req.DatabaseProvider)
	authorized := false
	for _, uid := range allowed {
		if uid == req.Token.UserId {
			authorized = true
			break
		}
	}
	if !authorized {
		return nil, http.StatusForbidden, errors.New("only a season commissioner may create weeks")
	}

	v, err := database.CreateOne(req.Context, req.DatabaseProvider, week)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: v}, http.StatusOK, nil
}

// WeekQuery holds the optional query parameters for ListWeeks.
type WeekQuery struct {
	// DraftId filters the weeks to a single Draft.
	DraftId model.DraftId `json:"draft_id"`
	// SeasonId filters the weeks to the Draft of a single Season.
	SeasonId model.SeasonId `json:"season_id"`
}

// StaticallyValid has no static constraints.
func (q *WeekQuery) StaticallyValid() error { return nil }

// ListWeeks returns all weeks, or filters them to a single Draft or Season via
// the `draft_id` / `season_id` query parameters. Weeks are accessible to
// everyone (they only carry a scheduled date and note), so no auth gate is
// applied.
type ListWeeks struct{}

func (c ListWeeks) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, BaseRoute
}

func (c ListWeeks) RequestBody() (*WeekQuery, bool) {
	return &WeekQuery{}, false
}

func (c ListWeeks) Handler(req api.Request[*WeekQuery]) (any, int, error) {
	query := req.HTTPRequest().URL.Query()

	if draftStr := query.Get("draft_id"); draftStr != "" {
		draftId, err := database.RecordIdFromString(draftStr)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		weeks, err := model.GetWeeksForDraft(req.Context, req.DatabaseProvider, model.DraftId(draftId))
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		return gin.H{api.ResourceKey: weeks}, http.StatusOK, nil
	}

	if seasonStr := query.Get("season_id"); seasonStr != "" {
		seasonId, err := database.RecordIdFromString(seasonStr)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		season, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, seasonId)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		weeks, err := model.GetWeeksForDraft(req.Context, req.DatabaseProvider, season.DraftId)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		return gin.H{api.ResourceKey: weeks}, http.StatusOK, nil
	}

	weeks, err := database.GetAll[*model.Week](req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: weeks}, http.StatusOK, nil
}
