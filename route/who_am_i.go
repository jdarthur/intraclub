package route

import (
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"
)

var WhoAmIRoute = "/whoami"

type WhoAmI struct{}

func (c WhoAmI) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, WhoAmIRoute
}

func (c WhoAmI) RequestBody() (*model.User, bool) {
	return &model.User{}, true
}

func (c WhoAmI) Handler(req api.Request[*model.User]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token must be passed into create user route")
	}

	user, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.User{}, req.Token.UserId.RecordId())
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}
	return user, http.StatusOK, nil
}
