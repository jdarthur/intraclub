package route

import (
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/model"
)

var GetMySeasonsRoute = "/seasons_composite"

type GetMySeasons struct{}

func (c GetMySeasons) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, GetMySeasonsRoute
}

func (c GetMySeasons) RequestBody() (*model.SeasonComposite, bool) {
	return &model.SeasonComposite{}, true
}

type seasonCompositeQueryParams struct {
	AsPlayer       []string `json:"as_player"`
	AsCommissioner []string `json:"as_commissioner"`
}

func (c GetMySeasons) Handler(req api.Request[*model.SeasonComposite]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token must be passed into create user route")
	}

	query := &seasonCompositeQueryParams{}
	err := req.ParseQuery(query)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	asPlayer := false
	if query.AsPlayer != nil && query.AsPlayer[0] == "true" {
		asPlayer = true
	}

	asCommissioner := false
	if query.AsCommissioner != nil && query.AsCommissioner[0] == "true" {
		asCommissioner = true
	}

	v, err := model.GetMySeasons(req.Context, req.DatabaseProvider, req.Token, asPlayer, asCommissioner)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return v, http.StatusOK, nil
}
