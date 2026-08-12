package organization

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base API route for Organization records. It matches the
// model's Type() ("organization").
var BaseRoute = "/organization"

// loadEditableOrganization fetches the Organization referenced by the request's
// path ID and verifies the requesting user (from the token) is authorized to
// edit it (owner / sysadmin). This mirrors the access-control pattern used by
// the Draft and Format custom routes.
func loadEditableOrganization[T database.Validatable](req api.Request[T]) (*model.Organization, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	org, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Organization{}, req.PathId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	wac := database.NewWithAccessControl[*model.Organization](req.Context, req.DatabaseProvider, req.Token.UserId)
	if !wac.CanUserEdit(org) {
		return nil, http.StatusForbidden, errors.New("not authorized to modify this organization")
	}

	return org, http.StatusOK, nil
}

// EmptyBody is used by routes that do not accept a request body.
type EmptyBody struct{}

func (b *EmptyBody) StaticallyValid() error {
	return nil
}

// AddMemberBody is the request body for AddMember.
type AddMemberBody struct {
	// UserId is the registered User to add to the organization's membership.
	UserId database.UserId `json:"user_id"`
}

// StaticallyValid ensures a user was provided.
func (b *AddMemberBody) StaticallyValid() error {
	if b.UserId == database.InvalidUserId {
		return errors.New("user_id must not be empty")
	}
	return nil
}

// ListMembers returns the current membership roster (a flat list of registered
// User records) for the organization referenced by the path ID. Any
// authenticated user may list members, since organizations are AccessibleTo
// everyone.
type ListMembers struct{}

func (c ListMembers) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, api.AppendPathId(BaseRoute) + "/members"
}

func (c ListMembers) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c ListMembers) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}

	// verify the organization exists before listing its members
	if _, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Organization{}, req.PathId); err != nil {
		return nil, http.StatusBadRequest, err
	}

	members, err := listMembers(req)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: members}, http.StatusOK, nil
}

// AddMember adds a registered User to the organization's membership roster.
// Only the organization owner (or a sysadmin) may add members. Adding the same
// user twice, or a non-existent user, is rejected.
type AddMember struct{}

func (c AddMember) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/members"
}

func (c AddMember) RequestBody() (*AddMemberBody, bool) {
	return &AddMemberBody{}, true
}

func (c AddMember) Handler(req api.Request[*AddMemberBody]) (any, int, error) {
	org, status, err := loadEditableOrganization(req)
	if err != nil {
		return nil, status, err
	}

	member := model.NewOrganizationMember()
	member.OrganizationId = org.ID
	member.UserId = req.Body.UserId

	created, err := database.CreateOne(req.Context, req.DatabaseProvider, member)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return gin.H{api.ResourceKey: created}, http.StatusOK, nil
}

// RemoveMember removes a User from the organization's membership roster. Only
// the organization owner (or a sysadmin) may remove members.
type RemoveMember struct{}

func (c RemoveMember) Path() (api.HttpMethod, string) {
	return api.HttpMethodDelete, api.AppendPathId(BaseRoute) + "/members/:userId"
}

func (c RemoveMember) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c RemoveMember) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	org, status, err := loadEditableOrganization(req)
	if err != nil {
		return nil, status, err
	}

	userId, err := userIdFromPath(req)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	membership, err := findMembership(req, org.ID, userId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if membership == nil {
		return nil, http.StatusNotFound, errors.New("user is not a member of this organization")
	}

	deleted, exists, err := database.DeleteOneById(req.Context, req.DatabaseProvider, membership, membership.GetId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if !exists {
		return nil, http.StatusNotFound, errors.New("user is not a member of this organization")
	}

	return gin.H{api.ResourceKey: deleted}, http.StatusOK, nil
}

// listMembers fetches the current membership roster of the organization and
// resolves each membership to its registered User record.
func listMembers[T database.Validatable](req api.Request[T]) ([]*model.User, error) {
	memberships, err := database.GetAllWhere[*model.OrganizationMember](req.Context, req.DatabaseProvider,
		func(_ context.Context, m *model.OrganizationMember) bool { return m.OrganizationId == model.OrganizationId(req.PathId) })
	if err != nil {
		return nil, err
	}

	users := make([]*model.User, 0, len(memberships))
	for _, m := range memberships {
		user, exists, err := database.GetOneById(req.Context, req.DatabaseProvider, &model.User{}, m.UserId.RecordId())
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

// findMembership looks up the OrganizationMember join record linking the given
// organization and user, returning nil if it does not exist.
func findMembership[T database.Validatable](req api.Request[T], orgID model.OrganizationId, userId database.UserId) (*model.OrganizationMember, error) {
	memberships, err := database.GetAllWhere[*model.OrganizationMember](req.Context, req.DatabaseProvider,
		func(_ context.Context, m *model.OrganizationMember) bool {
			return m.OrganizationId == orgID && m.UserId == userId
		})
	if err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return nil, nil
	}
	return memberships[0], nil
}

// userIdFromPath parses the :userId path parameter (the final segment of the
// request path) into a database.UserId.
func userIdFromPath[T database.Validatable](req api.Request[T]) (database.UserId, error) {
	path := req.HTTPRequest().URL.Path
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) == 0 {
		return database.InvalidUserId, errors.New("invalid :userId path parameter")
	}
	raw := segments[len(segments)-1]
	rid, err := database.RecordIdFromString(raw)
	if err != nil {
		return database.InvalidUserId, errors.New("invalid :userId path parameter")
	}
	return database.UserId(rid), nil
}
