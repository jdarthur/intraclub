package format

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for Format records.
var BaseRoute = "/format"

// SetPossibleRatingsBody is the request body for setting a Format's possible
// ratings. The `ratings` list replaces the format's entire set of assigned
// ratings (as FormatRating join records), preserving order.
type SetPossibleRatingsBody struct {
	Ratings model.RatingList `json:"ratings"`
}

// StaticallyValid ensures the request body has the fields required to set
// possible ratings.
func (b *SetPossibleRatingsBody) StaticallyValid() error {
	if len(b.Ratings) == 0 {
		return errors.New("ratings must not be empty")
	}
	return nil
}

// GetPossibleRatings returns a Format's currently-assigned ratings, ordered
// highest-skill to lowest-skill (by RatingIndex).
type GetPossibleRatings struct{}

func (c GetPossibleRatings) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, api.AppendPathId(BaseRoute) + "/possible_ratings"
}

func (c GetPossibleRatings) RequestBody() (*SetPossibleRatingsBody, bool) {
	return &SetPossibleRatingsBody{}, false
}

func (c GetPossibleRatings) Handler(req api.Request[*SetPossibleRatingsBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required to get possible ratings")
	}

	format, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Format{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	ratings, err := resolvePossibleRatings(req.Context, req.DatabaseProvider, format)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: ratings}, http.StatusOK, nil
}

// SetPossibleRatings replaces a Format's assigned ratings (the FormatRating
// join records) with the provided list. Only the format's owner (or a
// SysAdmin) may do so.
type SetPossibleRatings struct{}

func (c SetPossibleRatings) Path() (api.HttpMethod, string) {
	return api.HttpMethodPut, api.AppendPathId(BaseRoute) + "/possible_ratings"
}

func (c SetPossibleRatings) RequestBody() (*SetPossibleRatingsBody, bool) {
	return &SetPossibleRatingsBody{}, true
}

func (c SetPossibleRatings) Handler(req api.Request[*SetPossibleRatingsBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required to set possible ratings")
	}

	format, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Format{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// enforce per-record edit authorization (owner / sysadmin)
	wac := database.NewWithAccessControl[*model.Format](req.Context, req.DatabaseProvider, req.Token.UserId)
	if !wac.CanUserEdit(format) {
		return nil, http.StatusForbidden, errors.New("not authorized to set ratings for this format")
	}

	if err := format.SetPossibleRatings(req.Context, req.DatabaseProvider, req.Body.Ratings); err != nil {
		return nil, http.StatusBadRequest, err
	}

	ratings, err := resolvePossibleRatings(req.Context, req.DatabaseProvider, format)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return gin.H{api.ResourceKey: ratings}, http.StatusOK, nil
}

// resolvePossibleRatings fetches the full Rating records for a format's
// possible ratings, preserving order.
func resolvePossibleRatings(ctx context.Context, db database.Provider, format *model.Format) ([]model.Rating, error) {
	ids, err := format.GetPossibleRatings(ctx, db)
	if err != nil {
		return nil, err
	}

	ratings := make([]model.Rating, 0, len(ids))
	for _, id := range ids {
		r, err := database.GetExistingRecordById(ctx, db, &model.Rating{}, id.RecordId())
		if err != nil {
			return nil, err
		}
		ratings = append(ratings, *r)
	}
	return ratings, nil
}
