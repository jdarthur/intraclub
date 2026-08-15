// Package comment exposes the interactive write surface for Comment records
// (reactions) that the generic CRUD surface in main.go doesn't cover.
//
// Generic CRUD (registered in main.go) handles create/list/read/update/delete
// of comments and the comment_reaction child rows. The routes in this package
// handle reacting/unreacting to a comment, which is deduplicated per
// (comment, user, type) via model.Comment.React/Unreact and requires the
// acting user to be a participant of the comment's blurb's season.
package comment

import (
	"errors"
	"fmt"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for Comment records. It matches the model's
// Type() ("comment").
var BaseRoute = "/comment"

// ReactionBody is the request body for React and Unreact. Reaction is the
// human-readable reaction name (e.g. "Thumbs up", "Fire", "Heart") resolved
// against model.GetAllReactionTypes().
type ReactionBody struct {
	Reaction string `json:"reaction"`
}

// StaticallyValid ensures a reaction name was provided.
func (b *ReactionBody) StaticallyValid() error {
	if b.Reaction == "" {
		return errors.New("reaction must not be empty")
	}
	return nil
}

// applyReaction runs React or Unreact for the token user against the comment at
// req.PathId and returns the comment's updated reaction list.
func applyReaction(req api.Request[*ReactionBody], unreact bool) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	commentRec, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Comment{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// The acting user must be a participant of the comment's blurb's season
	// (mirrors the participant check enforced for blurb reactions).
	blurb, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Blurb{}, commentRec.Blurb.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := blurb.CanUserCommentOrReact(req.Context, req.DatabaseProvider, req.Token.UserId); err != nil {
		return nil, http.StatusForbidden, err
	}

	reactionTypes := model.GetAllReactionTypes()
	t, ok := reactionTypes[req.Body.Reaction]
	if !ok {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid reaction %q", req.Body.Reaction)
	}

	if unreact {
		err = commentRec.Unreact(req.Context, req.DatabaseProvider, req.Token.UserId, t)
	} else {
		err = commentRec.React(req.Context, req.DatabaseProvider, req.Token.UserId, t)
	}
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	reactions, err := commentRec.GetReactions(req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return gin.H{api.ResourceKey: reactions}, http.StatusOK, nil
}

// React adds a reaction from the token user to the comment.
type React struct{}

func (c React) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/react"
}

func (c React) RequestBody() (*ReactionBody, bool) {
	return &ReactionBody{}, true
}

func (c React) Handler(req api.Request[*ReactionBody]) (any, int, error) {
	return applyReaction(req, false)
}

// Unreact removes the token user's reaction of the given type from the comment.
type Unreact struct{}

func (c Unreact) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/unreact"
}

func (c Unreact) RequestBody() (*ReactionBody, bool) {
	return &ReactionBody{}, true
}

func (c Unreact) Handler(req api.Request[*ReactionBody]) (any, int, error) {
	return applyReaction(req, true)
}
