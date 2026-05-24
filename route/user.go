package route

import (
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"
)

var UserBaseRoute = "/user"

// SelfRegister allows a user to self-register to the system
type SelfRegister struct{}

func (c SelfRegister) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, UserBaseRoute
}

func (c SelfRegister) RequestBody() (*model.User, bool) {
	return &model.User{}, true
}

func (c SelfRegister) Handler(req api.Request[*model.User]) (any, int, error) {
	if req.Body.ID.RecordId() != database.InvalidRecordId {
		return nil, http.StatusBadRequest, errors.New("user ID must not be passed into create user route")
	}
	if req.Token != nil {
		return nil, http.StatusBadRequest, errors.New("token must not be passed into create user route")
	}

	user, err := req.DatabaseProvider.Create(req.Context, req.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return user, http.StatusCreated, nil
}
