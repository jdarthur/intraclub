// Package seasoncommissioner exposes the write surface for SeasonCommissioner
// records (adding/removing co-commissioners to a season).
//
// The generic CRUD surface in main.go is registered read-only for this type;
// writes go exclusively through the routes in this package, which enforce the
// model's sysadmin-only EditableBy constraint.
package seasoncommissioner

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for SeasonCommissioner records. It matches
// the model's Type() ("season_commissioner").
var BaseRoute = "/season_commissioner"

// AddCommissionerBody is the request body for AddCommissioner.
type AddCommissionerBody struct {
	// SeasonId is the season the user is being added as a co-commissioner to.
	SeasonId model.SeasonId `json:"season_id"`

	// UserId is the user being added as a co-commissioner.
	UserId database.UserId `json:"user_id"`
}

// StaticallyValid ensures both a season and a user were provided.
func (b *AddCommissionerBody) StaticallyValid() error {
	if b.SeasonId == model.SeasonId(database.InvalidRecordId) {
		return errors.New("season_id must not be empty")
	}
	if b.UserId == database.InvalidUserId {
		return errors.New("user_id must not be empty")
	}
	return nil
}

// AddCommissioner creates a SeasonCommissioner record linking the given user
// to the given season. Only a system administrator may add co-commissioners.
type AddCommissioner struct{}

func (c AddCommissioner) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute
}

func (c AddCommissioner) RequestBody() (*AddCommissionerBody, bool) {
	return &AddCommissionerBody{}, true
}

func (c AddCommissioner) Handler(req api.Request[*AddCommissionerBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	isAdmin, err := model.IsUserSystemAdministrator(req.Context, req.DatabaseProvider, req.Token.UserId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !isAdmin {
		return nil, http.StatusForbidden, errors.New("only a system administrator may add co-commissioners to a season")
	}

	season, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, req.Body.SeasonId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := season.AddCommissioner(req.Context, req.DatabaseProvider, req.Body.UserId); err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Return the created record so the client can use its ID to remove it.
	commissioners, err := database.GetAllWhere[*model.SeasonCommissioner](req.Context, req.DatabaseProvider,
		func(_ context.Context, c *model.SeasonCommissioner) bool {
			return c.SeasonId == season.ID && c.UserId == req.Body.UserId
		})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(commissioners) == 0 {
		return nil, http.StatusInternalServerError, errors.New("co-commissioner was not created")
	}
	return gin.H{api.ResourceKey: commissioners[0]}, http.StatusOK, nil
}

// RemoveCommissioner deletes the SeasonCommissioner record with the given ID.
// Only a system administrator may remove co-commissioners.
type RemoveCommissioner struct{}

func (c RemoveCommissioner) Path() (api.HttpMethod, string) {
	return api.HttpMethodDelete, api.AppendPathId(BaseRoute)
}

func (c RemoveCommissioner) RequestBody() (*AddCommissionerBody, bool) {
	return &AddCommissionerBody{}, false
}

func (c RemoveCommissioner) Handler(req api.Request[*AddCommissionerBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	isAdmin, err := model.IsUserSystemAdministrator(req.Context, req.DatabaseProvider, req.Token.UserId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !isAdmin {
		return nil, http.StatusForbidden, errors.New("only a system administrator may remove co-commissioners from a season")
	}

	existing, exists, err := database.GetOneById(req.Context, req.DatabaseProvider, &model.SeasonCommissioner{}, req.PathId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !exists {
		return nil, http.StatusNotFound, errors.New("co-commissioner not found")
	}
	if _, _, err := database.DeleteOneById(req.Context, req.DatabaseProvider, existing, existing.ID); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return gin.H{api.ResourceKey: existing}, http.StatusOK, nil
}
