// Package lateaddition exposes the write surface for SeasonLateAddition
// records (adding/removing players to a season after the draft completed).
//
// The generic CRUD surface in main.go is registered read-only for this type;
// writes go exclusively through the routes in this package, which enforce the
// model's sysadmin-only EditableBy constraint. (The generic create path does
// not check EditableBy, so a plain POST /api/season_late_addition would
// otherwise let any logged-in user add late players.)
package lateaddition

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for SeasonLateAddition records. It matches
// the model's Type() ("season_late_addition").
var BaseRoute = "/season_late_addition"

// AddLateAdditionBody is the request body for AddLateAddition.
type AddLateAdditionBody struct {
	// SeasonId is the season the user is being added to.
	SeasonId model.SeasonId `json:"season_id"`

	// UserId is the user being added to the season after the draft.
	UserId database.UserId `json:"user_id"`
}

// StaticallyValid ensures both a season and a user were provided.
func (b *AddLateAdditionBody) StaticallyValid() error {
	if b.SeasonId == model.SeasonId(database.InvalidRecordId) {
		return errors.New("season_id must not be empty")
	}
	if b.UserId == database.InvalidUserId {
		return errors.New("user_id must not be empty")
	}
	return nil
}

// AddLateAddition creates a SeasonLateAddition record linking the given user
// to the given season. Only a system administrator may add late players.
type AddLateAddition struct{}

func (c AddLateAddition) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute
}

func (c AddLateAddition) RequestBody() (*AddLateAdditionBody, bool) {
	return &AddLateAdditionBody{}, true
}

func (c AddLateAddition) Handler(req api.Request[*AddLateAdditionBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	isAdmin, err := model.IsUserSystemAdministrator(req.Context, req.DatabaseProvider, req.Token.UserId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !isAdmin {
		return nil, http.StatusForbidden, errors.New("only a system administrator may add late players to a season")
	}

	season, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, req.Body.SeasonId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := season.AddLateAddition(req.Context, req.DatabaseProvider, req.Body.UserId); err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Return the created record so the client can use its ID to remove it.
	additions, err := database.GetAllWhere[*model.SeasonLateAddition](req.Context, req.DatabaseProvider,
		func(_ context.Context, l *model.SeasonLateAddition) bool {
			return l.SeasonId == season.ID && l.UserId == req.Body.UserId
		})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(additions) == 0 {
		return nil, http.StatusInternalServerError, errors.New("late addition was not created")
	}
	return gin.H{api.ResourceKey: additions[0]}, http.StatusOK, nil
}

// RemoveLateAddition deletes the SeasonLateAddition record with the given ID.
// Only a system administrator may remove late players.
type RemoveLateAddition struct{}

func (c RemoveLateAddition) Path() (api.HttpMethod, string) {
	return api.HttpMethodDelete, api.AppendPathId(BaseRoute)
}

func (c RemoveLateAddition) RequestBody() (*AddLateAdditionBody, bool) {
	return &AddLateAdditionBody{}, false
}

func (c RemoveLateAddition) Handler(req api.Request[*AddLateAdditionBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	isAdmin, err := model.IsUserSystemAdministrator(req.Context, req.DatabaseProvider, req.Token.UserId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !isAdmin {
		return nil, http.StatusForbidden, errors.New("only a system administrator may remove late players from a season")
	}

	existing, exists, err := database.GetOneById(req.Context, req.DatabaseProvider, &model.SeasonLateAddition{}, req.PathId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !exists {
		return nil, http.StatusNotFound, errors.New("late addition not found")
	}
	if _, _, err := database.DeleteOneById(req.Context, req.DatabaseProvider, existing, existing.ID); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return gin.H{api.ResourceKey: existing}, http.StatusOK, nil
}
