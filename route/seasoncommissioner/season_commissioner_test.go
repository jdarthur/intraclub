package seasoncommissioner

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
// test setup helpers (mirror route/lateaddition/late_addition_test.go)
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

// newTestRouter builds a gin engine with the co-commissioner write surface
// registered (as in main.go) and the auth globals wired up so real tokens
// validate.
func newTestRouter(t *testing.T, db database.Provider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	api.UserType = &model.User{}
	database.SysAdminCheck = model.IsUserSystemAdministrator

	router := gin.New()
	group := router.Group("/api")

	commissioners := api.RouteFamily[*AddCommissionerBody]{DatabaseProvider: db}
	commissioners.Handle(group, AddCommissioner{}, RemoveCommissioner{})

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

// newSeason creates a minimal Season record (the fields required for
// DynamicallyValid on a SeasonCommissioner to pass).
func newSeason(t *testing.T, db database.Provider, owner *model.User) *model.Season {
	t.Helper()
	season := model.NewSeason()
	season.Name = fmt.Sprintf("Season %d", rand.Uint64())
	season.StartTime = model.NewStartTime(8, 30)
	season.Owner = owner.ID
	v, err := database.CreateOne(context.Background(), db, season)
	require.NoError(t, err)
	return v
}

func commissionerUserIds(t *testing.T, db database.Provider, seasonId model.SeasonId) []database.UserId {
	t.Helper()
	rows, err := database.GetAllWhere[*model.SeasonCommissioner](context.Background(), db,
		func(_ context.Context, c *model.SeasonCommissioner) bool { return c.SeasonId == seasonId })
	require.NoError(t, err)
	ids := make([]database.UserId, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.UserId)
	}
	return ids
}

// ---------------------------------------------------------------------------
// AddCommissioner
// ---------------------------------------------------------------------------

func TestAddCommissioner_SysAdmin(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	admin := newStoredUser(t, db)
	require.NoError(t, admin.AssignRole(context.Background(), db, model.SystemAdministrator))
	season := newSeason(t, db, admin)
	co := newStoredUser(t, db)
	token := newToken(t, admin.ID)

	w := doJSON(t, router, http.MethodPost, "/api/season_commissioner",
		AddCommissionerBody{SeasonId: season.ID, UserId: co.ID}, token)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var resp struct {
		Resource model.SeasonCommissioner `json:"resource"`
	}
	require.NoError(t, json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&resp))
	require.Equal(t, season.ID, resp.Resource.SeasonId)
	require.Equal(t, co.ID, resp.Resource.UserId)

	require.Contains(t, commissionerUserIds(t, db, season.ID), co.ID)
}

func TestAddCommissioner_NonSysAdminForbidden(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	regular := newStoredUser(t, db)
	season := newSeason(t, db, regular)
	other := newStoredUser(t, db)
	token := newToken(t, regular.ID)

	w := doJSON(t, router, http.MethodPost, "/api/season_commissioner",
		AddCommissionerBody{SeasonId: season.ID, UserId: other.ID}, token)
	require.Equal(t, http.StatusForbidden, w.Code, "body: %s", w.Body.String())

	require.Empty(t, commissionerUserIds(t, db, season.ID))
}

func TestAddCommissioner_NoToken(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	season := newSeason(t, db, newStoredUser(t, db))

	w := doJSON(t, router, http.MethodPost, "/api/season_commissioner",
		AddCommissionerBody{SeasonId: season.ID, UserId: newStoredUser(t, db).ID}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "body: %s", w.Body.String())
}

// ---------------------------------------------------------------------------
// RemoveCommissioner
// ---------------------------------------------------------------------------

func addCommissioner(t *testing.T, db database.Provider, season *model.Season, user *model.User) *model.SeasonCommissioner {
	t.Helper()
	require.NoError(t, season.AddCommissioner(context.Background(), db, user.ID))
	rows, err := database.GetAllWhere[*model.SeasonCommissioner](context.Background(), db,
		func(_ context.Context, c *model.SeasonCommissioner) bool {
			return c.SeasonId == season.ID && c.UserId == user.ID
		})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	return rows[0]
}

func TestRemoveCommissioner_SysAdmin(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	admin := newStoredUser(t, db)
	require.NoError(t, admin.AssignRole(context.Background(), db, model.SystemAdministrator))
	season := newSeason(t, db, admin)
	co := newStoredUser(t, db)
	sc := addCommissioner(t, db, season, co)
	token := newToken(t, admin.ID)

	w := doJSON(t, router, http.MethodDelete, fmt.Sprintf("/api/season_commissioner/%s", sc.ID), nil, token)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	require.NotContains(t, commissionerUserIds(t, db, season.ID), co.ID)
}

func TestRemoveCommissioner_NonSysAdminForbidden(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	admin := newStoredUser(t, db)
	require.NoError(t, admin.AssignRole(context.Background(), db, model.SystemAdministrator))
	season := newSeason(t, db, admin)
	co := newStoredUser(t, db)
	sc := addCommissioner(t, db, season, co)
	regular := newStoredUser(t, db)
	token := newToken(t, regular.ID)

	w := doJSON(t, router, http.MethodDelete, fmt.Sprintf("/api/season_commissioner/%s", sc.ID), nil, token)
	require.Equal(t, http.StatusForbidden, w.Code, "body: %s", w.Body.String())

	require.Contains(t, commissionerUserIds(t, db, season.ID), co.ID)
}

func TestRemoveCommissioner_NotFound(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	admin := newStoredUser(t, db)
	require.NoError(t, admin.AssignRole(context.Background(), db, model.SystemAdministrator))
	token := newToken(t, admin.ID)

	missing := database.NewRecordId()
	w := doJSON(t, router, http.MethodDelete, fmt.Sprintf("/api/season_commissioner/%s", missing.String()), nil, token)
	require.Equal(t, http.StatusNotFound, w.Code, "body: %s", w.Body.String())
}
