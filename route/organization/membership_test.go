package organization

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// test setup helpers (mirror route/draft/draft_test.go)
// ---------------------------------------------------------------------------

func newStoredUser(t *testing.T, db database.Provider) *model.User {
	t.Helper()
	user := model.NewUser()
	user.Email = model.EmailAddress(fmt.Sprintf("user%d@email.com", rand.Uint64()))
	user.FirstName = fmt.Sprintf("Test %d", rand.Uint64())
	user.LastName = "User"
	user.PhoneNumber = model.PhoneNumber(fmt.Sprintf("%d", 100_000_0000+rand.Uint32N(999_999_999)))
	v, err := database.CreateOne(context.Background(), db, user)
	require.NoError(t, err)
	return v
}

// newTestRouter builds a gin engine with the organization CRUD + membership
// surface registered (as in main.go) and the auth globals wired up so real
// tokens validate.
func newTestRouter(t *testing.T, db database.Provider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	api.UserType = &model.User{}
	database.SysAdminCheck = model.IsUserSystemAdministrator

	router := gin.New()
	group := router.Group("/api")

	orgs := api.NewCrudCommon(model.NewOrganization, false, db)
	orgs.HandleRouteTypes(group, api.CrudWrapperFunctionAll...)
	RegisterRoutes(group, db)

	return router
}

// testKeyOnce generates a single JWT keypair shared by every token minted in a
// test process, so tokens issued by different helpers all verify.
var (
	testKeyOnce sync.Once
	testPubKey  *ecdsa.PublicKey
	testPrivKey *ecdsa.PrivateKey
)

// newToken returns a signed JWT for the given user ID without touching key
// files on disk.
func newToken(t *testing.T, userId database.UserId) string {
	t.Helper()
	testKeyOnce.Do(func() {
		pub, priv, err := api.GenerateKeyPair()
		require.NoError(t, err)
		testPubKey = pub
		testPrivKey = priv
	})
	api.JwtPublicKey = testPubKey
	api.JwtPrivateKey = testPrivKey
	token, err := api.GenerateToken(userId.RecordId())
	require.NoError(t, err)
	return token
}

// doJSON performs an authenticated HTTP request against the router.
func doJSON(t *testing.T, router *gin.Engine, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, path, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set(api.AuthTokenHeaderValue, token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// createOrgViaHTTP creates an organization through the CRUD route as the given
// owner, returning its ID string.
func createOrgViaHTTP(t *testing.T, router *gin.Engine, owner database.UserId, name string) string {
	t.Helper()
	w := doJSON(t, router, http.MethodPost, "/api/organization", map[string]any{
		"name": name,
	}, newToken(t, owner))
	require.Equal(t, http.StatusOK, w.Code, "create org: %s", w.Body.String())

	var resp struct {
		Resource *model.Organization `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotNil(t, resp.Resource)
	require.Equal(t, owner, resp.Resource.GetOwner())
	return resp.Resource.ID.String()
}

// ---------------------------------------------------------------------------
// Organization CRUD
// ---------------------------------------------------------------------------

func TestOrganizationCRUD(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	owner := newStoredUser(t, db)

	orgID := createOrgViaHTTP(t, router, owner.ID, "Martin's Landing Men")

	// GET one
	w := doJSON(t, router, http.MethodGet, "/api/organization/"+orgID, nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var one struct {
		Resource *model.Organization `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &one))
	require.Equal(t, orgID, one.Resource.ID.String())
	require.Equal(t, owner.ID, one.Resource.GetOwner())

	// GET many
	w = doJSON(t, router, http.MethodGet, "/api/organization", nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var many struct {
		Resource []*model.Organization `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &many))
	require.Len(t, many.Resource, 1)

	// PUT update
	w = doJSON(t, router, http.MethodPut, "/api/organization/"+orgID, map[string]any{
		"id":   orgID,
		"name": "Renamed Org",
	}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "update org: %s", w.Body.String())

	// DELETE
	w = doJSON(t, router, http.MethodDelete, "/api/organization/"+orgID, nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code)

	w = doJSON(t, router, http.MethodGet, "/api/organization/"+orgID, nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusNotFound, w.Code)
}

// ---------------------------------------------------------------------------
// Membership endpoints
// ---------------------------------------------------------------------------

func TestMembershipAddListRemove(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	owner := newStoredUser(t, db)
	orgID := createOrgViaHTTP(t, router, owner.ID, "Members Org")
	memberA := newStoredUser(t, db)
	memberB := newStoredUser(t, db)

	// add memberA
	w := doJSON(t, router, http.MethodPost, "/api/organization/"+orgID+"/members",
		map[string]any{"user_id": memberA.ID.String()}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "add member: %s", w.Body.String())
	var added struct {
		Resource *model.OrganizationMember `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &added))
	require.NotNil(t, added.Resource)
	require.Equal(t, memberA.ID, added.Resource.UserId)

	// add memberB
	w = doJSON(t, router, http.MethodPost, "/api/organization/"+orgID+"/members",
		map[string]any{"user_id": memberB.ID.String()}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "add member B: %s", w.Body.String())

	// list members (any authenticated user)
	w = doJSON(t, router, http.MethodGet, "/api/organization/"+orgID+"/members", nil, newToken(t, memberA.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var listed struct {
		Resource []*model.User `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.Len(t, listed.Resource, 2)

	// remove memberA
	w = doJSON(t, router, http.MethodDelete, "/api/organization/"+orgID+"/members/"+memberA.ID.String(), nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "remove member: %s", w.Body.String())

	w = doJSON(t, router, http.MethodGet, "/api/organization/"+orgID+"/members", nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.Len(t, listed.Resource, 1)
	require.Equal(t, memberB.ID, listed.Resource[0].ID)
}

func TestMembershipAuthz(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	owner := newStoredUser(t, db)
	orgID := createOrgViaHTTP(t, router, owner.ID, "Authz Org")
	nonOwner := newStoredUser(t, db)
	member := newStoredUser(t, db)

	// non-owner cannot add
	w := doJSON(t, router, http.MethodPost, "/api/organization/"+orgID+"/members",
		map[string]any{"user_id": member.ID.String()}, newToken(t, nonOwner.ID))
	require.Equal(t, http.StatusForbidden, w.Code)

	// non-owner cannot remove (even though member was never added, authz check
	// happens before membership lookup)
	w = doJSON(t, router, http.MethodDelete, "/api/organization/"+orgID+"/members/"+member.ID.String(), nil, newToken(t, nonOwner.ID))
	require.Equal(t, http.StatusForbidden, w.Code)

	// unauthenticated cannot list
	w = doJSON(t, router, http.MethodGet, "/api/organization/"+orgID+"/members", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMembershipRejectsDuplicateAndUnknown(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	owner := newStoredUser(t, db)
	orgID := createOrgViaHTTP(t, router, owner.ID, "Dup Org")
	member := newStoredUser(t, db)

	// unknown user rejected
	unknownID := database.NewRecordId()
	w := doJSON(t, router, http.MethodPost, "/api/organization/"+orgID+"/members",
		map[string]any{"user_id": unknownID.String()}, newToken(t, owner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "unknown user: %s", w.Body.String())

	// add member once
	w = doJSON(t, router, http.MethodPost, "/api/organization/"+orgID+"/members",
		map[string]any{"user_id": member.ID.String()}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code)

	// duplicate add rejected
	w = doJSON(t, router, http.MethodPost, "/api/organization/"+orgID+"/members",
		map[string]any{"user_id": member.ID.String()}, newToken(t, owner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "duplicate: %s", w.Body.String())

	// removing a non-member returns 404
	w = doJSON(t, router, http.MethodDelete, "/api/organization/"+orgID+"/members/"+member.ID.String(), nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code)

	w = doJSON(t, router, http.MethodDelete, "/api/organization/"+orgID+"/members/"+member.ID.String(), nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusNotFound, w.Code)
}
