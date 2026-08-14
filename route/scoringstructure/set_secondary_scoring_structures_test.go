package scoringstructure

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

// newSysAdminUser creates a user and assigns them the SystemAdministrator
// role directly (the raw scoring_structure_secondary CRUD surface is
// sysadmin-only).
func newSysAdminUser(t *testing.T, db database.Provider) *model.User {
	t.Helper()
	user := newStoredUser(t, db)
	assignment := &model.UserRoleAssignment{
		UserId: user.ID,
		Role:   model.SystemAdministrator,
	}
	_, err := database.CreateOne(context.Background(), db, assignment)
	require.NoError(t, err)
	return user
}

// newTestRouter builds a gin engine with the scoring structure + secondary
// CRUD surface and the get/set secondary routes registered (as in main.go)
// with the auth globals wired up so real tokens validate.
func newTestRouter(t *testing.T, db database.Provider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	api.UserType = &model.User{}
	database.SysAdminCheck = model.IsUserSystemAdministrator

	router := gin.New()
	group := router.Group("/api")

	scoringStructures := api.NewCrudCommon(model.NewScoringStructure, false, db)
	scoringStructures.HandleRouteTypes(group, api.CrudWrapperFunctionAll...)

	scoringStructureSecondaries := api.NewCrudCommon(func() *model.ScoringStructureSecondary { return &model.ScoringStructureSecondary{} }, false, db)
	scoringStructureSecondaries.HandleRouteTypes(group, api.CrudWrapperFunctionAll...)

	family := api.RouteFamily[*SetSecondaryScoringStructuresBody]{DatabaseProvider: db}
	family.Handle(group, GetSecondaryScoringStructures{}, SetSecondaryScoringStructures{})

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

// createScoringStructureViaHTTP creates a scoring structure through the CRUD
// route as the given owner, returning its ID string.
func createScoringStructureViaHTTP(t *testing.T, router *gin.Engine, owner database.UserId, name string, countingType model.ScoreCountingType, winThreshold, mustWinBy, instantWin int) string {
	t.Helper()
	w := doJSON(t, router, http.MethodPost, "/api/scoring_structure", map[string]any{
		"name":                        name,
		"win_condition_counting_type": countingType,
		"win_condition": map[string]any{
			"win_threshold":         winThreshold,
			"must_win_by":           mustWinBy,
			"instant_win_threshold": instantWin,
		},
	}, newToken(t, owner))
	require.Equal(t, http.StatusOK, w.Code, "create scoring structure: %s", w.Body.String())

	var resp struct {
		Resource *model.ScoringStructure `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotNil(t, resp.Resource)
	return resp.Resource.ID.String()
}

// getSecondaries performs GET /api/scoring_structure/:id/secondary_scoring_structures
// and returns the ordered list of secondary IDs.
func getSecondaries(t *testing.T, router *gin.Engine, structureID, token string) []string {
	t.Helper()
	w := doJSON(t, router, http.MethodGet, "/api/scoring_structure/"+structureID+"/secondary_scoring_structures", nil, token)
	require.Equal(t, http.StatusOK, w.Code, "get secondaries: %s", w.Body.String())

	var resp struct {
		Resource []model.ScoringStructure `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	ids := make([]string, 0, len(resp.Resource))
	for _, s := range resp.Resource {
		ids = append(ids, s.ID.String())
	}
	return ids
}

func setSecondaries(t *testing.T, router *gin.Engine, structureID string, ids []string, token string) *httptest.ResponseRecorder {
	t.Helper()
	return doJSON(t, router, http.MethodPut, "/api/scoring_structure/"+structureID+"/secondary_scoring_structures",
		map[string]any{"secondary_scoring_structures": ids}, token)
}

// ---------------------------------------------------------------------------
// Get / set secondary scoring structures
// ---------------------------------------------------------------------------

func TestSecondaryScoringStructuresSetGetReorderClear(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	owner := newStoredUser(t, db)

	// Set-based primary: best-of-3 sets (max 3 units) needs game-based secondaries.
	primary := createScoringStructureViaHTTP(t, router, owner.ID, "Primary match", model.Set, 2, 1, 0)
	gameA := createScoringStructureViaHTTP(t, router, owner.ID, "Game A", model.Game, 6, 2, 7)
	gameB := createScoringStructureViaHTTP(t, router, owner.ID, "Game B", model.Game, 6, 2, 7)

	// No secondaries yet.
	require.Empty(t, getSecondaries(t, router, primary, newToken(t, owner.ID)))

	// Set an ordered list (duplicates at different positions are allowed).
	w := setSecondaries(t, router, primary, []string{gameA, gameB, gameA}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "set secondaries: %s", w.Body.String())
	require.Equal(t, []string{gameA, gameB, gameA}, getSecondaries(t, router, primary, newToken(t, owner.ID)))

	// Reorder: the list is replaced atomically and the new order is preserved.
	w = setSecondaries(t, router, primary, []string{gameB, gameA, gameA}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "reorder secondaries: %s", w.Body.String())
	require.Equal(t, []string{gameB, gameA, gameA}, getSecondaries(t, router, primary, newToken(t, owner.ID)))

	// Clearing the list makes the structure non-composite again.
	w = setSecondaries(t, router, primary, []string{}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "clear secondaries: %s", w.Body.String())
	require.Empty(t, getSecondaries(t, router, primary, newToken(t, owner.ID)))
}

func TestSecondaryScoringStructuresAuthz(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	owner := newStoredUser(t, db)
	nonOwner := newStoredUser(t, db)

	primary := createScoringStructureViaHTTP(t, router, owner.ID, "Authz match", model.Set, 2, 1, 0)
	gameA := createScoringStructureViaHTTP(t, router, owner.ID, "Authz game", model.Game, 6, 2, 7)

	// non-owner cannot set
	w := setSecondaries(t, router, primary, []string{gameA}, newToken(t, nonOwner.ID))
	require.Equal(t, http.StatusForbidden, w.Code)

	// non-owner can still read (scoring structures are accessible to everyone)
	require.Equal(t, []string{}, getSecondaries(t, router, primary, newToken(t, nonOwner.ID)))

	// unauthenticated cannot get or set
	w = doJSON(t, router, http.MethodGet, "/api/scoring_structure/"+primary+"/secondary_scoring_structures", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
	w = setSecondaries(t, router, primary, []string{gameA}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// a sysadmin who is not the owner can set
	sysadmin := newSysAdminUser(t, db)
	w = setSecondaries(t, router, primary, []string{gameA}, newToken(t, sysadmin.ID))
	require.Equal(t, http.StatusOK, w.Code, "sysadmin set secondaries: %s", w.Body.String())
	require.Equal(t, []string{gameA}, getSecondaries(t, router, primary, newToken(t, owner.ID)))
}

func TestSecondaryScoringStructuresValidation(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	owner := newStoredUser(t, db)

	primary := createScoringStructureViaHTTP(t, router, owner.ID, "Validation match", model.Set, 2, 1, 0)
	gameA := createScoringStructureViaHTTP(t, router, owner.ID, "Validation game", model.Game, 6, 2, 7)
	pointA := createScoringStructureViaHTTP(t, router, owner.ID, "Validation point", model.Point, 11, 1, 0)

	// unknown secondary ID rejected
	w := setSecondaries(t, router, primary, []string{database.NewRecordId().String()}, newToken(t, owner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "unknown secondary: %s", w.Body.String())

	// wrong counting type rejected (point-based secondary on a set-based primary)
	w = setSecondaries(t, router, primary, []string{pointA}, newToken(t, owner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "wrong counting type: %s", w.Body.String())
	require.Empty(t, getSecondaries(t, router, primary, newToken(t, owner.ID)))

	// a valid set still works afterwards
	w = setSecondaries(t, router, primary, []string{gameA}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "valid set: %s", w.Body.String())
	require.Equal(t, []string{gameA}, getSecondaries(t, router, primary, newToken(t, owner.ID)))
}

// ---------------------------------------------------------------------------
// Raw scoring_structure_secondary CRUD (sysadmin-only writes)
// ---------------------------------------------------------------------------

func TestScoringStructureSecondaryCRUD(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	owner := newStoredUser(t, db)
	sysadmin := newSysAdminUser(t, db)

	primary := createScoringStructureViaHTTP(t, router, owner.ID, "CRUD match", model.Set, 2, 1, 0)
	gameA := createScoringStructureViaHTTP(t, router, owner.ID, "CRUD game", model.Game, 6, 2, 7)

	// generic CRUD create is open to any authenticated user (the record's own
	// validation runs; EditableBy gates update / delete)
	w := doJSON(t, router, http.MethodPost, "/api/scoring_structure_secondary", map[string]any{
		"scoring_structure_id":           primary,
		"secondary_scoring_structure_id": gameA,
		"secondary_index":                0,
	}, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code, "create: %s", w.Body.String())

	var created struct {
		Resource *model.ScoringStructureSecondary `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotNil(t, created.Resource)
	rowID := created.Resource.ID.String()

	// everyone can list the raw collection
	w = doJSON(t, router, http.MethodGet, "/api/scoring_structure_secondary", nil, newToken(t, owner.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var listed struct {
		Resource []*model.ScoringStructureSecondary `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.Len(t, listed.Resource, 1)

	// duplicate (primary, index) rejected
	w = doJSON(t, router, http.MethodPost, "/api/scoring_structure_secondary", map[string]any{
		"scoring_structure_id":           primary,
		"secondary_scoring_structure_id": gameA,
		"secondary_index":                0,
	}, newToken(t, sysadmin.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "duplicate index: %s", w.Body.String())

	// unknown primary rejected
	w = doJSON(t, router, http.MethodPost, "/api/scoring_structure_secondary", map[string]any{
		"scoring_structure_id":           database.NewRecordId().String(),
		"secondary_scoring_structure_id": gameA,
		"secondary_index":                0,
	}, newToken(t, sysadmin.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "unknown primary: %s", w.Body.String())

	// a non-sysadmin cannot update the raw join row
	w = doJSON(t, router, http.MethodPut, "/api/scoring_structure_secondary/"+rowID, map[string]any{
		"id":                             rowID,
		"scoring_structure_id":           primary,
		"secondary_scoring_structure_id": gameA,
		"secondary_index":                1,
	}, newToken(t, owner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "non-sysadmin update: %s", w.Body.String())

	// the sysadmin can update and delete the raw join row
	w = doJSON(t, router, http.MethodPut, "/api/scoring_structure_secondary/"+rowID, map[string]any{
		"id":                             rowID,
		"scoring_structure_id":           primary,
		"secondary_scoring_structure_id": gameA,
		"secondary_index":                1,
	}, newToken(t, sysadmin.ID))
	require.Equal(t, http.StatusOK, w.Code, "sysadmin update: %s", w.Body.String())

	w = doJSON(t, router, http.MethodDelete, "/api/scoring_structure_secondary/"+rowID, nil, newToken(t, sysadmin.ID))
	require.Equal(t, http.StatusOK, w.Code, "sysadmin delete: %s", w.Body.String())

	w = doJSON(t, router, http.MethodGet, "/api/scoring_structure_secondary/"+rowID, nil, newToken(t, sysadmin.ID))
	require.Equal(t, http.StatusNotFound, w.Code)
}
