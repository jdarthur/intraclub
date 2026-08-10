package ruleset

import (
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for Ruleset records.
var BaseRoute = "/ruleset"

// AmendNameBody is the request body for amending a Ruleset's name.
type AmendNameBody struct {
	Name string `json:"name"`
}

// StaticallyValid ensures the request body has the fields required to amend a Ruleset.
func (b *AmendNameBody) StaticallyValid() error {
	if b.Name == "" {
		return errors.New("name must not be empty")
	}
	return nil
}

// AmendRulesetName is a custom route that amends a Ruleset by producing a new
// revision with a new name (see model.Ruleset.EditName). Rulesets may not be
// directly modified (Ruleset.PreUpdate forbids it), so an update is expressed
// as an amend that creates a superseding revision.
type AmendRulesetName struct{}

func (c AmendRulesetName) Path() (api.HttpMethod, string) {
	return api.HttpMethodPut, api.AppendPathId(BaseRoute) + "/amend"
}

func (c AmendRulesetName) RequestBody() (*AmendNameBody, bool) {
	return &AmendNameBody{}, true
}

func (c AmendRulesetName) Handler(req api.Request[*AmendNameBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required to amend a ruleset")
	}

	ruleset, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Ruleset{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// enforce per-record edit authorization (owner / sysadmin)
	wac := database.NewWithAccessControl[*model.Ruleset](req.Context, req.DatabaseProvider, req.Token.UserId)
	if !wac.CanUserEdit(ruleset) {
		return nil, http.StatusForbidden, errors.New("not authorized to amend this ruleset")
	}

	newRuleset, err := ruleset.EditName(req.Context, req.DatabaseProvider, req.Body.Name)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: newRuleset}, http.StatusOK, nil
}
