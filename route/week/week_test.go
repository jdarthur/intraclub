package week

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
	"time"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// test helpers (mirror route/draft and model test helpers)
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

func newStoredRating(t *testing.T, db database.Provider) *model.Rating {
	t.Helper()
	user := newStoredUser(t, db)
	r := model.NewRating()
	r.UserId = user.ID
	r.Name = fmt.Sprintf("Rating %s", database.NewRecordId())
	r.Description = "test description"
	v, err := database.CreateOne(context.Background(), db, r)
	require.NoError(t, err)
	return v
}

func newDefaultFormat(t *testing.T, db database.Provider) *model.Format {
	t.Helper()
	user := newStoredUser(t, db)
	f := model.NewFormat()
	f.UserId = user.ID
	f.Name = "default format"
	created, err := database.CreateOne(context.Background(), db, f)
	require.NoError(t, err)

	lines := []model.FormatLine{
		{Player1Rating: newStoredRating(t, db).ID, Player2Rating: newStoredRating(t, db).ID},
		{Player1Rating: newStoredRating(t, db).ID, Player2Rating: newStoredRating(t, db).ID},
	}
	var ratings model.RatingList
	for _, l := range lines {
		ratings = appendUniqueRating(ratings, l.Player1Rating)
		ratings = appendUniqueRating(ratings, l.Player2Rating)
	}
	require.NoError(t, created.SetPossibleRatings(context.Background(), db, ratings))
	require.NoError(t, created.SetLines(context.Background(), db, lines))
	return created
}

func appendUniqueRating(ratings model.RatingList, r model.RatingId) model.RatingList {
	for _, existing := range ratings {
		if existing == r {
			return ratings
		}
	}
	return append(ratings, r)
}

// newSeasonWithTeam builds a Season (with a commissioner and a draft) plus one
// team whose captain is returned. The team is assigned to the season, making
// the captain a "participant" who must not be able to create/modify weeks.
func newSeasonWithTeam(t *testing.T, db database.Provider, commissionerID database.UserId) (*model.Season, *model.User) {
	t.Helper()
	ctx := context.Background()

	format := newDefaultFormat(t, db)
	draft := model.NewDraft()
	draft.Owner = commissionerID
	draft.Format = format.ID
	draftV, err := database.CreateOne(ctx, db, draft)
	require.NoError(t, err)

	facility := model.NewFacility()
	facility.UserId = commissionerID
	facility.Name = "Test facility"
	facility.Address = "Test Rd."
	facility.NumberOfCourts = 2
	facilityV, err := database.CreateOne(ctx, db, facility)
	require.NoError(t, err)

	season := model.NewSeason()
	season.Name = "Test Season"
	season.StartTime = model.NewStartTime(8, 30)
	season.DraftId = draftV.ID
	season.Facility = facilityV.ID
	seasonV, err := database.CreateOne(ctx, db, season)
	require.NoError(t, err)
	require.NoError(t, seasonV.AddCommissioner(ctx, db, commissionerID))

	captain := newStoredUser(t, db)
	team := model.NewDefaultTeam(captain.ID, "Team A")
	teamV, err := database.CreateOne(ctx, db, team)
	require.NoError(t, err)
	require.NoError(t, seasonV.AddTeam(ctx, db, teamV.ID))

	return seasonV, captain
}

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

// ---------------------------------------------------------------------------
// commissioner-only create
// ---------------------------------------------------------------------------

func createWeekViaHTTP(t *testing.T, router *gin.Engine, draftID string, token string) *httptest.ResponseRecorder {
	t.Helper()
	return doJSON(t, router, http.MethodPost, "/api/week", map[string]any{
		"draft_id": draftID,
		"date":     time.Date(2025, 3, 1, 8, 0, 0, 0, time.UTC),
		"note":     "week 1",
	}, token)
}

func TestCreateWeekCommissionerOnly(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, captain := newSeasonWithTeam(t, db, commissioner.ID)
	outsider := newStoredUser(t, db)

	draftID := season.DraftId.String()

	// Commissioner may create a week.
	w := createWeekViaHTTP(t, router, draftID, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "commissioner create: %s", w.Body.String())
	var created struct {
		Resource *model.Week `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotNil(t, created.Resource)
	require.Equal(t, season.DraftId, created.Resource.DraftId)

	// A team captain (season participant) may NOT create a week.
	w = createWeekViaHTTP(t, router, draftID, newToken(t, captain.ID))
	require.Equal(t, http.StatusForbidden, w.Code, "captain create: %s", w.Body.String())

	// An unrelated user (normal participant) may NOT create a week.
	w = createWeekViaHTTP(t, router, draftID, newToken(t, outsider.ID))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider create: %s", w.Body.String())

	// No token -> unauthorized.
	w = createWeekViaHTTP(t, router, draftID, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeekUpdateDeleteCommissionerOnly(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, captain := newSeasonWithTeam(t, db, commissioner.ID)

	// Create a week as the commissioner.
	w := createWeekViaHTTP(t, router, season.DraftId.String(), newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var created struct {
		Resource *model.Week `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	weekID := created.Resource.ID.String()

	// The captain cannot update the week's date.
	w = doJSON(t, router, http.MethodPut, "/api/week/"+weekID, map[string]any{
		"id":      weekID,
		"draft_id": season.DraftId.String(),
		"date":    time.Date(2025, 3, 8, 8, 0, 0, 0, time.UTC),
		"note":    "moved",
	}, newToken(t, captain.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "captain update: %s", w.Body.String())

	// The commissioner can update it.
	w = doJSON(t, router, http.MethodPut, "/api/week/"+weekID, map[string]any{
		"id":       weekID,
		"draft_id": season.DraftId.String(),
		"date":     time.Date(2025, 3, 8, 8, 0, 0, 0, time.UTC),
		"note":     "moved",
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "commissioner update: %s", w.Body.String())

	// The captain cannot delete the week: the delete is rejected by access
	// control (CanUserEdit) and the week remains.
	w = doJSON(t, router, http.MethodDelete, "/api/week/"+weekID, nil, newToken(t, captain.ID))
	require.Equal(t, http.StatusOK, w.Code, "captain delete: %s", w.Body.String())
	w = doJSON(t, router, http.MethodGet, "/api/week/"+weekID, nil, newToken(t, captain.ID))
	require.Equal(t, http.StatusOK, w.Code)

	// An outsider cannot delete it either.
	outsider := newStoredUser(t, db)
	w = doJSON(t, router, http.MethodDelete, "/api/week/"+weekID, nil, newToken(t, outsider.ID))
	require.Equal(t, http.StatusOK, w.Code, "outsider delete: %s", w.Body.String())
	w = doJSON(t, router, http.MethodGet, "/api/week/"+weekID, nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// query by draft / season
// ---------------------------------------------------------------------------

func TestWeeksQueryByDraftAndSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, _ := newSeasonWithTeam(t, db, commissioner.ID)

	createWeekViaHTTP(t, router, season.DraftId.String(), newToken(t, commissioner.ID))
	createWeekViaHTTP(t, router, season.DraftId.String(), newToken(t, commissioner.ID))

	// Query by draft_id.
	w := doJSON(t, router, http.MethodGet, "/api/week?draft_id="+season.DraftId.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var byDraft struct {
		Resource []*model.Week `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &byDraft))
	require.Len(t, byDraft.Resource, 2)

	// Query by season_id.
	w = doJSON(t, router, http.MethodGet, "/api/week?season_id="+season.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var bySeason struct {
		Resource []*model.Week `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &bySeason))
	require.Len(t, bySeason.Resource, 2)

	// Unfiltered list returns all weeks.
	w = doJSON(t, router, http.MethodGet, "/api/week", nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var all struct {
		Resource []*model.Week `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &all))
	require.Len(t, all.Resource, 2)
}

// ---------------------------------------------------------------------------
// body validation
// ---------------------------------------------------------------------------

func TestCreateWeekBodyValidation(t *testing.T) {
	b := &CreateWeekBody{}
	require.Error(t, b.StaticallyValid())
	b.Date = time.Now()
	require.NoError(t, b.StaticallyValid())
}
