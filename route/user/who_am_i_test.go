package user

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
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
// test setup helpers (mirror route/team/promote_test.go)
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

// newTestRouter builds a gin engine with the whoami + roles endpoints
// registered (as in main.go) and the auth globals wired up so real tokens
// validate.
func newTestRouter(t *testing.T, db database.Provider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	api.UserType = &model.User{}
	database.SysAdminCheck = model.IsUserSystemAdministrator

	router := gin.New()
	group := router.Group("/api")

	whoAmI := api.RouteFamily[*model.User]{DatabaseProvider: db}
	whoAmI.Handle(group, WhoAmI{})

	roles := api.RouteFamily[*model.User]{DatabaseProvider: db}
	roles.Handle(group, Roles{})

	return router
}

// testKeyOnce generates a single JWT keypair shared by every token minted in a
// test process, so tokens issued by different helpers all verify.
var (
	testKeyOnce sync.Once
	testPubKey  *ecdsa.PublicKey
	testPrivKey *ecdsa.PrivateKey
)

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

// doGet performs an authenticated HTTP GET against the router.
func doGet(t *testing.T, router *gin.Engine, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set(api.AuthTokenHeaderValue, token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// WhoAmI
// ---------------------------------------------------------------------------

// TestWhoAmI returns 200 and the caller's user for a valid token. Regression
// for #196: RequestBody previously declared a body, so the generic wrapper
// StaticallyValid'd a zero-valued User and aborted with 400 before the handler
// ran.
func TestWhoAmI_ValidToken(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	user := newStoredUser(t, db)
	token := newToken(t, user.ID)

	w := doGet(t, router, "/api/whoami", token)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var resp model.User
	err := json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, user.ID, resp.ID)
	require.Equal(t, user.FirstName, resp.FirstName)
}

// TestWhoAmI_NoToken returns 401 (not 400) when no token is supplied.
func TestWhoAmI_NoToken(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)

	w := doGet(t, router, "/api/whoami", "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "body: %s", w.Body.String())
}

// ---------------------------------------------------------------------------
// Roles
// ---------------------------------------------------------------------------

// TestRoles returns the caller's role names for a valid token.
func TestRoles_ValidToken(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	user := newStoredUser(t, db)
	require.NoError(t, user.AssignRole(context.Background(), db, model.SystemAdministrator))
	token := newToken(t, user.ID)

	w := doGet(t, router, "/api/whoami/roles", token)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var resp struct {
		Roles []string `json:"roles"`
	}
	err := json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&resp)
	require.NoError(t, err)
	require.Contains(t, resp.Roles, model.SystemAdministrator.String())
}

// TestRoles_NoToken returns 401 when no token is supplied.
func TestRoles_NoToken(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)

	w := doGet(t, router, "/api/whoami/roles", "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "body: %s", w.Body.String())
}
