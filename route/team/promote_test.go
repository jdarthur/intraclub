package team

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
// test setup helpers (mirror route/organization/membership_test.go)
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

// newTestRouter builds a gin engine with the Team read-only surface + the
// promote endpoint registered (as in main.go) and the auth globals wired up so
// real tokens validate.
func newTestRouter(t *testing.T, db database.Provider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	api.UserType = &model.User{}
	database.SysAdminCheck = model.IsUserSystemAdministrator

	router := gin.New()
	group := router.Group("/api")

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

// newStoredTeam creates a Team record with the given user as captain (as the
// draft finalize flow would) and returns it.
func newStoredTeam(t *testing.T, db database.Provider, captain *model.User) *model.Team {
	t.Helper()
	team := model.NewDefaultTeam(captain.ID, "Team 1")
	team, err := database.CreateOne(context.Background(), db, team)
	require.NoError(t, err)
	return team
}

// addAssignment adds a TeamAssignment row for the given user on the team.
func addAssignment(t *testing.T, db database.Provider, team *model.Team, user *model.User, role model.TeamRole) *model.TeamAssignment {
	t.Helper()
	a := &model.TeamAssignment{TeamId: team.ID, UserId: user.ID, Role: role}
	v, err := database.CreateOne(context.Background(), db, a)
	require.NoError(t, err)
	return v
}

func assignmentRole(t *testing.T, db database.Provider, team *model.Team, user *model.User) model.TeamRole {
	t.Helper()
	assignments, err := database.GetAllWhere[*model.TeamAssignment](context.Background(), db,
		func(_ context.Context, a *model.TeamAssignment) bool { return a.TeamId == team.ID && a.UserId == user.ID })
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	return assignments[0].Role
}

// ---------------------------------------------------------------------------
// Promote endpoint
// ---------------------------------------------------------------------------

func TestPromoteCoCaptain(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	captain := newStoredUser(t, db)
	member := newStoredUser(t, db)

	team := newStoredTeam(t, db, captain)
	addAssignment(t, db, team, member, model.TeamRoleMember)

	// The captain promotes the member.
	w := doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{"user_id": member.ID.String()}, newToken(t, captain.ID))
	require.Equal(t, http.StatusOK, w.Code, "promote: %s", w.Body.String())

	var promoted struct {
		Resource *model.TeamAssignment `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &promoted))
	require.NotNil(t, promoted.Resource)
	require.Equal(t, model.TeamRoleCoCaptain, promoted.Resource.Role)
	require.Equal(t, model.TeamRoleCoCaptain, assignmentRole(t, db, team, member))
}

func TestPromoteCoCaptainByCoCaptain(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	captain := newStoredUser(t, db)
	coCaptain := newStoredUser(t, db)
	member := newStoredUser(t, db)

	team := newStoredTeam(t, db, captain)
	addAssignment(t, db, team, coCaptain, model.TeamRoleCoCaptain)
	addAssignment(t, db, team, member, model.TeamRoleMember)

	// A co-captain may also assign roles.
	w := doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{"user_id": member.ID.String()}, newToken(t, coCaptain.ID))
	require.Equal(t, http.StatusOK, w.Code, "promote by co-captain: %s", w.Body.String())
	require.Equal(t, model.TeamRoleCoCaptain, assignmentRole(t, db, team, member))
}

func TestPromoteCoCaptainAuthz(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	captain := newStoredUser(t, db)
	member := newStoredUser(t, db)
	outsider := newStoredUser(t, db)

	team := newStoredTeam(t, db, captain)
	addAssignment(t, db, team, member, model.TeamRoleMember)

	// A non-member cannot promote.
	w := doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{"user_id": member.ID.String()}, newToken(t, outsider.ID))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider promote: %s", w.Body.String())
	require.Equal(t, model.TeamRoleMember, assignmentRole(t, db, team, member))

	// Unauthenticated request is rejected.
	w = doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{"user_id": member.ID.String()}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPromoteCoCaptainInvalidCases(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	captain := newStoredUser(t, db)
	coCaptain := newStoredUser(t, db)
	member := newStoredUser(t, db)
	nonMember := newStoredUser(t, db)

	team := newStoredTeam(t, db, captain)
	addAssignment(t, db, team, coCaptain, model.TeamRoleCoCaptain)
	addAssignment(t, db, team, member, model.TeamRoleMember)

	captainToken := newToken(t, captain.ID)

	// Cannot promote the captain.
	w := doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{"user_id": captain.ID.String()}, captainToken)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// Cannot promote an existing co-captain.
	w = doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{"user_id": coCaptain.ID.String()}, captainToken)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// Cannot promote a user who is not on the team.
	w = doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{"user_id": nonMember.ID.String()}, captainToken)
	require.Equal(t, http.StatusNotFound, w.Code)

	// Missing user_id is rejected statically.
	w = doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{}, captainToken)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// Unknown team id.
	w = doJSON(t, router, http.MethodPost, "/api/team/999/promote_co_captain",
		map[string]any{"user_id": member.ID.String()}, captainToken)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// Commissioner read access
// ---------------------------------------------------------------------------

func TestSeasonCommissionerCanViewTeam(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	captain := newStoredUser(t, db)
	member := newStoredUser(t, db)
	commissioner := newStoredUser(t, db)

	team := newStoredTeam(t, db, captain)
	addAssignment(t, db, team, member, model.TeamRoleMember)

	// Wire the team into a season with the commissioner attached.
	season := model.NewSeason()
	season.Name = "Test Season"
	season.StartTime = model.NewStartTime(8, 30)
	season, err := database.CreateOne(context.Background(), db, season)
	require.NoError(t, err)

	seasonTeam := &model.SeasonTeam{SeasonId: season.ID, TeamId: team.ID}
	_, err = database.CreateOne(context.Background(), db, seasonTeam)
	require.NoError(t, err)

	seasonCommissioner := &model.SeasonCommissioner{SeasonId: season.ID, UserId: commissioner.ID}
	_, err = database.CreateOne(context.Background(), db, seasonCommissioner)
	require.NoError(t, err)

	// The commissioner can view the team (list + detail) ...
	w := doJSON(t, router, http.MethodGet, "/api/team/"+team.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "commissioner get team: %s", w.Body.String())

	w = doJSON(t, router, http.MethodGet, "/api/team", nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var many struct {
		Resource []TeamRoster `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &many))
	require.Len(t, many.Resource, 1)

	// ... but cannot promote a member (role assignment is a team function).
	w = doJSON(t, router, http.MethodPost, "/api/team/"+team.ID.String()+"/promote_co_captain",
		map[string]any{"user_id": member.ID.String()}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusForbidden, w.Code, "commissioner promote: %s", w.Body.String())
	require.Equal(t, model.TeamRoleMember, assignmentRole(t, db, team, member))
}

// ---------------------------------------------------------------------------
// Read-only surface: no generic create/update/delete on raw records
// ---------------------------------------------------------------------------

func TestTeamReadSurface(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	captain := newStoredUser(t, db)
	member := newStoredUser(t, db)
	outsider := newStoredUser(t, db)

	team := newStoredTeam(t, db, captain)
	addAssignment(t, db, team, member, model.TeamRoleMember)

	// GET one + GET many work for a team member.
	w := doJSON(t, router, http.MethodGet, "/api/team/"+team.ID.String(), nil, newToken(t, captain.ID))
	require.Equal(t, http.StatusOK, w.Code, "get team: %s", w.Body.String())
	var one struct {
		Resource TeamRoster `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &one))
	require.Equal(t, team.ID, one.Resource.Team.ID)
	require.Len(t, one.Resource.Assignments, 2)

	w = doJSON(t, router, http.MethodGet, "/api/team", nil, newToken(t, captain.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var many struct {
		Resource []TeamRoster `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &many))
	require.Len(t, many.Resource, 1)

	// A non-member cannot see the team.
	w = doJSON(t, router, http.MethodGet, "/api/team", nil, newToken(t, outsider.ID))
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &many))
	require.Empty(t, many.Resource)

	w = doJSON(t, router, http.MethodGet, "/api/team/"+team.ID.String(), nil, newToken(t, outsider.ID))
	require.Equal(t, http.StatusNotFound, w.Code)

	// Raw create/update/delete on Team and TeamAssignment are not registered:
	// they must not be routable.
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		path := "/api/team"
		if method != http.MethodPost {
			path += "/" + team.ID.String()
		}
		w = doJSON(t, router, method, path, map[string]any{"name": "x"}, newToken(t, captain.ID))
		require.Equal(t, http.StatusNotFound, w.Code, "%s %s should not be routable", method, path)

		path = "/api/team_assignment"
		if method != http.MethodPost {
			path += "/1"
		}
		w = doJSON(t, router, method, path, map[string]any{}, newToken(t, captain.ID))
		require.Equal(t, http.StatusNotFound, w.Code, "%s %s should not be routable", method, path)
	}
}
