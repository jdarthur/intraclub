package user

import (
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// RolesRoute is the path for the current user's role list.
var RolesRoute = "/whoami/roles"

// Roles returns the role names of the authenticated user, e.g.
// `{"roles": ["System Administrator"]}`. The UI uses it to gate
// sysadmin-only controls (such as season late additions).
type Roles struct{}

func (c Roles) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, RolesRoute
}

func (c Roles) RequestBody() (*model.User, bool) {
	// roles takes no request body: the caller's identity comes from the token.
	// Declaring a body here makes the generic wrapper bind + StaticallyValid a
	// zero-valued User and 400 before the handler runs (see #196).
	return nil, false
}

func (c Roles) Handler(req api.Request[*model.User]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	user, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.User{}, req.Token.UserId.RecordId())
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	roles, err := user.GetRoles(req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return gin.H{"roles": roles}, http.StatusOK, nil
}
