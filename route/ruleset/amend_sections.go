package ruleset

import (
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// AmendSections applies a RuleAmendment (add / remove / modify / reorder a
// section) to a Ruleset. Rulesets may not be directly modified
// (Ruleset.PreUpdate forbids it), so section edits go through the
// Ruleset.Amend flow: add/remove/reorder produce a new superseding revision,
// while a pure content edit updates the section in place. Returns the (possibly
// new) Ruleset revision.
type AmendSections struct{}

func (c AmendSections) Path() (api.HttpMethod, string) {
	return api.HttpMethodPut, api.AppendPathId(BaseRoute) + "/amend_sections"
}

func (c AmendSections) RequestBody() (*model.RuleAmendment, bool) {
	return &model.RuleAmendment{}, true
}

func (c AmendSections) Handler(req api.Request[*model.RuleAmendment]) (any, int, error) {
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

	newRuleset, err := ruleset.Amend(req.Context, req.DatabaseProvider, req.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: newRuleset}, http.StatusOK, nil
}
