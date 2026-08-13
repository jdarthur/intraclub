package availability

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
// test setup helpers (mirror route/week and route/team helpers)
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

// newSeasonWithTeam builds a Season (with a commissioner and a draft) plus a
// team whose captain is returned. A co-captain and a regular member are added
// as assignments, and a single week is created for the draft. All four users
// are season participants (the captain/co-captain/member via team membership,
// the commissioner via the commissioner join).
type seasonFixture struct {
	season       *model.Season
	draftId      model.DraftId
	team         *model.Team
	week         *model.Week
	captain      database.UserId
	coCaptain    database.UserId
	member       database.UserId
	commissioner database.UserId
}

func newSeasonWithTeam(t *testing.T, db database.Provider) *seasonFixture {
	t.Helper()
	ctx := context.Background()

	commissioner := newStoredUser(t, db)

	draft := model.NewDraft()
	draft.Owner = commissioner.ID
	draft.Format = model.FormatId(newStoredFormat(t, db).ID)
	draftV, err := database.CreateOne(ctx, db, draft)
	require.NoError(t, err)

	facility := model.NewFacility()
	facility.UserId = commissioner.ID
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
	require.NoError(t, seasonV.AddCommissioner(ctx, db, commissioner.ID))

	captain := newStoredUser(t, db)
	team := model.NewDefaultTeam(captain.ID, "Team A")
	teamV, err := database.CreateOne(ctx, db, team)
	require.NoError(t, err)
	require.NoError(t, seasonV.AddTeam(ctx, db, teamV.ID))

	coCaptain := newStoredUser(t, db)
	member := newStoredUser(t, db)
	addAssignment(t, db, teamV, coCaptain, model.TeamRoleCoCaptain)
	addAssignment(t, db, teamV, member, model.TeamRoleMember)

	week := model.NewWeek()
	week.DraftId = draftV.ID
	week.Date = time.Date(2025, 3, 1, 8, 0, 0, 0, time.UTC)
	week.Note = "week 1"
	weekV, err := database.CreateOne(ctx, db, week)
	require.NoError(t, err)

	return &seasonFixture{
		season:       seasonV,
		draftId:      draftV.ID,
		team:         teamV,
		week:         weekV,
		captain:      captain.ID,
		coCaptain:    coCaptain.ID,
		member:       member.ID,
		commissioner: commissioner.ID,
	}
}

func newStoredFormat(t *testing.T, db database.Provider) *model.Format {
	t.Helper()
	user := newStoredUser(t, db)
	f := model.NewFormat()
	f.UserId = user.ID
	f.Name = fmt.Sprintf("format %d", rand.Uint64())
	v, err := database.CreateOne(context.Background(), db, f)
	require.NoError(t, err)
	return v
}

func addAssignment(t *testing.T, db database.Provider, team *model.Team, user *model.User, role model.TeamRole) *model.TeamAssignment {
	t.Helper()
	a := &model.TeamAssignment{TeamId: team.ID, UserId: user.ID, Role: role}
	v, err := database.CreateOne(context.Background(), db, a)
	require.NoError(t, err)
	return v
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
// set availability
// ---------------------------------------------------------------------------

func TestSetAvailabilityCreatesAndUpserts(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newSeasonWithTeam(t, db)

	// A participant sets availability for the week.
	w := doJSON(t, router, http.MethodPost, "/api/availability/set",
		map[string]any{"week_id": fx.week.ID.String(), "available": 1}, newToken(t, fx.member))
	require.Equal(t, http.StatusOK, w.Code, "set: %s", w.Body.String())
	var created struct {
		Resource *model.Availability `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotNil(t, created.Resource)
	require.Equal(t, fx.member, created.Resource.UserId)
	require.Equal(t, fx.week.ID, created.Resource.WeekId)
	require.Equal(t, model.AvailabilityAvailable, created.Resource.Available)

	// Setting the same week again updates (upserts) instead of duplicating.
	w = doJSON(t, router, http.MethodPost, "/api/availability/set",
		map[string]any{"week_id": fx.week.ID.String(), "available": 3}, newToken(t, fx.member))
	require.Equal(t, http.StatusOK, w.Code, "re-set: %s", w.Body.String())
	var updated struct {
		Resource *model.Availability `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	require.Equal(t, model.AvailabilityNotAvailable, updated.Resource.Available)
	require.Equal(t, created.Resource.ID, updated.Resource.ID, "expected the same record to be updated, not a new one")

	// Exactly one record exists for this user+week.
	all, err := database.GetAllWhere[*model.Availability](context.Background(), db,
		func(_ context.Context, a *model.Availability) bool {
			return a.UserId == fx.member && a.WeekId == fx.week.ID
		})
	require.NoError(t, err)
	require.Len(t, all, 1)
}

func TestSetAvailabilityRejectsNonParticipant(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newSeasonWithTeam(t, db)
	outsider := newStoredUser(t, db)

	w := doJSON(t, router, http.MethodPost, "/api/availability/set",
		map[string]any{"week_id": fx.week.ID.String(), "available": 1}, newToken(t, outsider.ID))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider set: %s", w.Body.String())

	// No token -> unauthorized.
	w = doJSON(t, router, http.MethodPost, "/api/availability/set",
		map[string]any{"week_id": fx.week.ID.String(), "available": 1}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSetAvailabilityRejectsInvalidOption(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newSeasonWithTeam(t, db)

	w := doJSON(t, router, http.MethodPost, "/api/availability/set",
		map[string]any{"week_id": fx.week.ID.String(), "available": 99}, newToken(t, fx.member))
	require.Equal(t, http.StatusBadRequest, w.Code, "invalid option: %s", w.Body.String())
}

// ---------------------------------------------------------------------------
// duplicate user+week prevention
// ---------------------------------------------------------------------------

func TestAvailabilityDuplicateUserWeekRejected(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newSeasonWithTeam(t, db)

	// Generic create of the same user+week twice is rejected by the model's
	// uniqueness constraint.
	body := map[string]any{"user_id": fx.member.String(), "week_id": fx.week.ID.String(), "available": 1}
	token := newToken(t, fx.member)
	w := doJSON(t, router, http.MethodPost, "/api/availability", body, token)
	require.Equal(t, http.StatusOK, w.Code, "first create: %s", w.Body.String())

	w = doJSON(t, router, http.MethodPost, "/api/availability", body, token)
	require.Equal(t, http.StatusBadRequest, w.Code, "second create: %s", w.Body.String())
}

// ---------------------------------------------------------------------------
// per-user query
// ---------------------------------------------------------------------------

func TestGetForUserReturnsOwnAvailability(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newSeasonWithTeam(t, db)

	doJSON(t, router, http.MethodPost, "/api/availability/set",
		map[string]any{"week_id": fx.week.ID.String(), "available": 1}, newToken(t, fx.member))

	// Member queries their own availability.
	w := doJSON(t, router, http.MethodGet, "/api/availability?draft_id="+fx.draftId.String(), nil, newToken(t, fx.member))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var mine struct {
		Resource []*model.Availability `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &mine))
	require.Len(t, mine.Resource, 1)
	require.Equal(t, fx.member, mine.Resource[0].UserId)

	// A different, non-sysadmin user cannot view another user's availability.
	w = doJSON(t, router, http.MethodGet,
		"/api/availability?draft_id="+fx.draftId.String()+"&user_id="+fx.member.String(),
		nil, newToken(t, fx.captain))
	require.Equal(t, http.StatusForbidden, w.Code, "view another user's: %s", w.Body.String())

	// The owner may view their own availability even with an explicit user_id.
	w = doJSON(t, router, http.MethodGet,
		"/api/availability?draft_id="+fx.draftId.String()+"&user_id="+fx.member.String(),
		nil, newToken(t, fx.member))
	require.Equal(t, http.StatusOK, w.Code, "owner self view: %s", w.Body.String())
}

// ---------------------------------------------------------------------------
// per-team query
// ---------------------------------------------------------------------------

func TestGetForTeamCaptainAndCoCaptain(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newSeasonWithTeam(t, db)

	// Captain, co-captain, and member each set availability.
	for _, uid := range []database.UserId{fx.captain, fx.coCaptain, fx.member} {
		w := doJSON(t, router, http.MethodPost, "/api/availability/set",
			map[string]any{"week_id": fx.week.ID.String(), "available": 1}, newToken(t, uid))
		require.Equal(t, http.StatusOK, w.Code, "set for %s: %s", uid, w.Body.String())
	}

	path := "/api/availability/team?team_id=" + fx.team.ID.String() + "&draft_id=" + fx.draftId.String()

	// Captain can view the team's availability.
	w := doJSON(t, router, http.MethodGet, path, nil, newToken(t, fx.captain))
	require.Equal(t, http.StatusOK, w.Code, "captain view: %s", w.Body.String())
	var capView struct {
		Resource []*TeamAvailabilityEntry `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &capView))
	require.Len(t, capView.Resource, 3, "expected availability for captain, co-captain, and member")
	for _, entry := range capView.Resource {
		require.Len(t, entry.Availabilities, 1)
		require.Equal(t, model.AvailabilityAvailable, entry.Availabilities[0].Available)
	}

	// Co-captain can view it too.
	w = doJSON(t, router, http.MethodGet, path, nil, newToken(t, fx.coCaptain))
	require.Equal(t, http.StatusOK, w.Code, "co-captain view: %s", w.Body.String())

	// A regular member cannot view the team's availability.
	w = doJSON(t, router, http.MethodGet, path, nil, newToken(t, fx.member))
	require.Equal(t, http.StatusForbidden, w.Code, "member view: %s", w.Body.String())

	// An outsider cannot view it either.
	outsider := newStoredUser(t, db)
	w = doJSON(t, router, http.MethodGet, path, nil, newToken(t, outsider.ID))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider view: %s", w.Body.String())

	// No token -> unauthorized.
	w = doJSON(t, router, http.MethodGet, path, nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// body validation
// ---------------------------------------------------------------------------

func TestSetAvailabilityBodyValidation(t *testing.T) {
	b := &SetAvailabilityBody{}
	b.Available = model.AvailabilityUnset
	require.NoError(t, b.StaticallyValid(), "unset is a valid option")

	b.Available = model.AvailabilityOption(99)
	require.Error(t, b.StaticallyValid(), "invalid option should be rejected")
}
