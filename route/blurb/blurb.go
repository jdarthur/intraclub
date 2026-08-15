// Package blurb exposes the interactive write surface for Blurb records that
// the generic CRUD surface in main.go doesn't cover: reacting to a blurb and
// attaching/detaching photos.
//
// Generic CRUD (registered in main.go) handles create/list/read/update/delete
// of blurbs, comments, and the child join tables. The routes in this package
// handle the operations that require business logic:
//
//   - React / Unreact: deduplicated per (blurb, user, type) via
//     model.Blurb.React/Unreact, which also enforces that the acting user is a
//     participant of the blurb's season.
//   - AddPhoto / RemovePhoto: the child BlurbPhoto.EditableBy is sysadmin-only,
//     so the normal owner flow goes through these routes, which enforce that
//     only the blurb owner may attach photos they own.
package blurb

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for Blurb records. It matches the model's
// Type() ("blurb").
var BaseRoute = "/blurb"

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

// applyReaction runs React or Unreact for the token user against the blurb at
// req.PathId and returns the blurb's updated reaction list.
func applyReaction(req api.Request[*ReactionBody], unreact bool) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	blurb, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Blurb{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// The acting user must be a participant of the blurb's season. Blurb.React
	// re-checks this internally, but we surface it here as a clean 403 so the
	// client can distinguish "forbidden" from an invalid request.
	if err := blurb.CanUserCommentOrReact(req.Context, req.DatabaseProvider, req.Token.UserId); err != nil {
		return nil, http.StatusForbidden, err
	}

	reactionTypes := model.GetAllReactionTypes()
	t, ok := reactionTypes[req.Body.Reaction]
	if !ok {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid reaction %q", req.Body.Reaction)
	}

	if unreact {
		err = blurb.Unreact(req.Context, req.DatabaseProvider, req.Token.UserId, t)
	} else {
		err = blurb.React(req.Context, req.DatabaseProvider, req.Token.UserId, t)
	}
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	reactions, err := blurb.GetReactions(req.Context, req.DatabaseProvider)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return gin.H{api.ResourceKey: reactions}, http.StatusOK, nil
}

// React adds a reaction from the token user to the blurb.
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

// Unreact removes the token user's reaction of the given type from the blurb.
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

// PhotoBody is the request body for AddPhoto.
type PhotoBody struct {
	PhotoId model.PhotoId `json:"photo_id"`
}

// StaticallyValid ensures a photo was provided.
func (b *PhotoBody) StaticallyValid() error {
	if b.PhotoId == model.PhotoId(database.InvalidRecordId) {
		return errors.New("photo_id must not be empty")
	}
	return nil
}

// AddPhoto attaches a photo owned by the token user to a blurb owned by the
// token user.
type AddPhoto struct{}

func (c AddPhoto) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/photos"
}

func (c AddPhoto) RequestBody() (*PhotoBody, bool) {
	return &PhotoBody{}, true
}

func (c AddPhoto) Handler(req api.Request[*PhotoBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	blurb, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Blurb{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if blurb.Owner != req.Token.UserId {
		return nil, http.StatusForbidden, errors.New("only the blurb owner may attach photos")
	}

	photo, exists, err := database.GetOneById(req.Context, req.DatabaseProvider, &model.Photo{}, req.Body.PhotoId.RecordId())
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !exists {
		return nil, http.StatusBadRequest, errors.New("photo does not exist")
	}
	if photo.Owner != req.Token.UserId {
		return nil, http.StatusForbidden, errors.New("photo must be owned by the blurb owner")
	}

	existing, err := database.GetAllWhere[*model.BlurbPhoto](req.Context, req.DatabaseProvider, func(_ context.Context, p *model.BlurbPhoto) bool {
		return p.BlurbId == blurb.ID && p.PhotoId == req.Body.PhotoId
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(existing) > 0 {
		return nil, http.StatusBadRequest, errors.New("photo is already attached to this blurb")
	}

	row := &model.BlurbPhoto{BlurbId: blurb.ID, PhotoId: req.Body.PhotoId}
	created, err := database.CreateOne(req.Context, req.DatabaseProvider, row)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: created}, http.StatusOK, nil
}

// photoIdFromPath parses the final :photoId path segment into a model.PhotoId.
func photoIdFromPath[T database.Validatable](req api.Request[T]) (model.PhotoId, error) {
	path := req.HTTPRequest().URL.Path
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) == 0 {
		return model.PhotoId(database.InvalidRecordId), errors.New("invalid :photoId path parameter")
	}
	raw := segments[len(segments)-1]
	rid, err := database.RecordIdFromString(raw)
	if err != nil {
		return model.PhotoId(database.InvalidRecordId), errors.New("invalid :photoId path parameter")
	}
	return model.PhotoId(rid), nil
}

// RemovePhoto detaches a photo from a blurb. Only the blurb owner may do so.
type RemovePhoto struct{}

func (c RemovePhoto) Path() (api.HttpMethod, string) {
	return api.HttpMethodDelete, api.AppendPathId(BaseRoute) + "/photos/" + ":photoId"
}

func (c RemovePhoto) RequestBody() (*PhotoBody, bool) {
	return &PhotoBody{}, false
}

func (c RemovePhoto) Handler(req api.Request[*PhotoBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	blurb, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Blurb{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if blurb.Owner != req.Token.UserId {
		return nil, http.StatusForbidden, errors.New("only the blurb owner may remove photos")
	}

	photoId, err := photoIdFromPath(req)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	rows, err := database.GetAllWhere[*model.BlurbPhoto](req.Context, req.DatabaseProvider, func(_ context.Context, p *model.BlurbPhoto) bool {
		return p.BlurbId == blurb.ID && p.PhotoId == photoId
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(rows) == 0 {
		return nil, http.StatusNotFound, errors.New("photo is not attached to this blurb")
	}
	for _, row := range rows {
		if _, _, err := database.DeleteOneById(req.Context, req.DatabaseProvider, row, row.ID); err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}
	return gin.H{api.ResourceKey: rows[0]}, http.StatusOK, nil
}
