package lineup

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
// test setup helpers (mirror route/availability)
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

// lineupFixture builds the chain required to build a lineup: a season (with a
// commissioner and a draft) whose format has a single line (rating1 / rating2),
// a team whose captain carries rating1 and whose member carries rating2, and a
// single week. This mirrors how a real drafted team ends up with rated members.
type lineupFixture struct {
	season       *model.Season
	week         *model.Week
	team         *model.Team
	line         model.FormatLine
	captain      database.UserId
	member       database.UserId
	commissioner database.UserId
	outsider     database.UserId
}

func newLineupFixture(t *testing.T, db database.Provider) *lineupFixture {
	t.Helper()
	ctx := context.Background()

	commissioner := newStoredUser(t, db)

	// A format with a single line (rating1 / rating2).
	rating1 := newStoredRating(t, db, commissioner.ID)
	rating2 := newStoredRating(t, db, commissioner.ID)
	format := model.NewFormat()
	format.UserId = commissioner.ID
	format.Name = fmt.Sprintf("format %d", rand.Uint64())
	formatV, err := database.CreateOne(ctx, db, format)
	require.NoError(t, err)
	line := model.FormatLine{Player1Rating: rating1, Player2Rating: rating2}
	require.NoError(t, formatV.SetPossibleRatings(ctx, db, model.RatingList{rating1, rating2}))
	require.NoError(t, formatV.SetLines(ctx, db, []model.FormatLine{line}))

	draft := model.NewDraft()
	draft.Owner = commissioner.ID
	draft.Format = formatV.ID
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
	member := newStoredUser(t, db)
	outsider := newStoredUser(t, db)
	team := model.NewDefaultTeam(captain.ID, "Team A")
	teamV, err := database.CreateOne(ctx, db, team)
	require.NoError(t, err)
	require.NoError(t, seasonV.AddTeam(ctx, db, teamV.ID))
	addAssignment(t, db, teamV, member, model.TeamRoleMember)

	// Assign the line's required ratings to the captain and member.
	_, err = database.CreateOne(ctx, db, &model.TeamRating{TeamId: teamV.ID, UserId: captain.ID, RatingId: rating1})
	require.NoError(t, err)
	_, err = database.CreateOne(ctx, db, &model.TeamRating{TeamId: teamV.ID, UserId: member.ID, RatingId: rating2})
	require.NoError(t, err)

	week := model.NewWeek()
	week.DraftId = draftV.ID
	week.Date = time.Date(2025, 3, 1, 8, 0, 0, 0, time.UTC)
	week.Note = "week 1"
	weekV, err := database.CreateOne(ctx, db, week)
	require.NoError(t, err)

	return &lineupFixture{
		season:       seasonV,
		week:         weekV,
		team:         teamV,
		line:         line,
		captain:      captain.ID,
		member:       member.ID,
		commissioner: commissioner.ID,
		outsider:     outsider.ID,
	}
}

func newStoredRating(t *testing.T, db database.Provider, owner database.UserId) model.RatingId {
	t.Helper()
	r := model.NewRating()
	r.UserId = owner
	r.Name = fmt.Sprintf("rating %d", rand.Uint64())
	r.Description = "test rating"
	v, err := database.CreateOne(context.Background(), db, r)
	require.NoError(t, err)
	return v.ID
}

func addAssignment(t *testing.T, db database.Provider, team *model.Team, user *model.User, role model.TeamRole) {
	t.Helper()
	_, err := database.CreateOne(context.Background(), db, &model.TeamAssignment{TeamId: team.ID, UserId: user.ID, Role: role})
	require.NoError(t, err)
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
	lineupKeyOnce sync.Once
	lineupPubKey  *ecdsa.PublicKey
	lineupPrivKey *ecdsa.PrivateKey
)

func newToken(t *testing.T, userId database.UserId) string {
	t.Helper()
	lineupKeyOnce.Do(func() {
		pub, priv, err := api.GenerateKeyPair()
		require.NoError(t, err)
		lineupPubKey = pub
		lineupPrivKey = priv
	})
	api.JwtPublicKey = lineupPubKey
	api.JwtPrivateKey = lineupPrivKey
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
// tests
// ---------------------------------------------------------------------------

func TestSetLineupAsCaptain(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newLineupFixture(t, db)

	// A non-captain (regular member) cannot build the lineup.
	w := doJSON(t, router, http.MethodPost, "/api/lineup/set",
		map[string]any{
			"team_id": fx.team.ID.String(),
			"week_id": fx.week.ID.String(),
			"pairings": []map[string]any{{
				"player1": fx.captain.String(),
				"player2": fx.member.String(),
				"format_line_index": 0,
			}},
		}, newToken(t, fx.member))
	require.Equal(t, http.StatusForbidden, w.Code, "member set: %s", w.Body.String())

	// The captain builds the lineup with one valid pairing.
	w = doJSON(t, router, http.MethodPost, "/api/lineup/set",
		map[string]any{
			"team_id": fx.team.ID.String(),
			"week_id": fx.week.ID.String(),
			"pairings": []map[string]any{{
				"player1": fx.captain.String(),
				"player2": fx.member.String(),
				"format_line_index": 0,
			}},
		}, newToken(t, fx.captain))
	require.Equal(t, http.StatusOK, w.Code, "captain set: %s", w.Body.String())

	var created struct {
		Resource *LineupDetail `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotNil(t, created.Resource)
	require.NotNil(t, created.Resource.Lineup)
	require.Equal(t, fx.team.ID, created.Resource.TeamId)
	require.Equal(t, fx.week.ID, created.Resource.WeekId)
	require.False(t, created.Resource.Lineup.Confirmed)
	require.Len(t, created.Resource.Pairings, 1)
	require.Equal(t, fx.captain, created.Resource.Pairings[0].Player1)
	require.Equal(t, fx.member, created.Resource.Pairings[0].Player2)
	require.Equal(t, 0, created.Resource.Pairings[0].FormatLineIndex)

	lineupID := created.Resource.Lineup.ID
	lineupIDStr := lineupID.String()

	// Confirm as the captain.
	w = doJSON(t, router, http.MethodPost, "/api/lineup/"+lineupIDStr+"/confirm", nil, newToken(t, fx.captain))
	require.Equal(t, http.StatusOK, w.Code, "confirm: %s", w.Body.String())
	var confirmed struct {
		Resource *model.Lineup `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &confirmed))
	require.True(t, confirmed.Resource.Confirmed)
	require.False(t, confirmed.Resource.Official)

	// A non-commissioner (the captain) cannot mark the lineup official.
	w = doJSON(t, router, http.MethodPost, "/api/lineup/"+lineupIDStr+"/official", nil, newToken(t, fx.captain))
	require.Equal(t, http.StatusForbidden, w.Code, "captain official: %s", w.Body.String())

	// The commissioner marks the confirmed lineup official.
	w = doJSON(t, router, http.MethodPost, "/api/lineup/"+lineupIDStr+"/official", nil, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, "official: %s", w.Body.String())
	var official struct {
		Resource *model.Lineup `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &official))
	require.True(t, official.Resource.Confirmed)
	require.True(t, official.Resource.Official)

	// GetLineupDetail reflects the saved state.
	w = doJSON(t, router, http.MethodGet,
		"/api/lineup/detail?team_id="+fx.team.ID.String()+"&week_id="+fx.week.ID.String(),
		nil, newToken(t, fx.member))
	require.Equal(t, http.StatusOK, w.Code, "detail: %s", w.Body.String())
	var detail struct {
		Resource *LineupDetail `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &detail))
	require.Len(t, detail.Resource.Lines, 1)
	require.Len(t, detail.Resource.Pairings, 1)
	require.True(t, detail.Resource.Lineup.Official)
}

func TestSetLineupRejectsMismatchedRatings(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newLineupFixture(t, db)

	// Swap the players so neither matches the line's required ratings.
	w := doJSON(t, router, http.MethodPost, "/api/lineup/set",
		map[string]any{
			"team_id": fx.team.ID.String(),
			"week_id": fx.week.ID.String(),
			"pairings": []map[string]any{{
				"player1": fx.member.String(),
				"player2": fx.captain.String(),
				"format_line_index": 0,
			}},
		}, newToken(t, fx.captain))
	require.Equal(t, http.StatusBadRequest, w.Code, "mismatch: %s", w.Body.String())
}

func TestConfirmRequiresCaptain(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newLineupFixture(t, db)

	// Build a lineup as the captain.
	w := doJSON(t, router, http.MethodPost, "/api/lineup/set",
		map[string]any{
			"team_id": fx.team.ID.String(),
			"week_id": fx.week.ID.String(),
			"pairings": []map[string]any{{
				"player1": fx.captain.String(),
				"player2": fx.member.String(),
				"format_line_index": 0,
			}},
		}, newToken(t, fx.captain))
	require.Equal(t, http.StatusOK, w.Code, "set: %s", w.Body.String())
	var created struct {
		Resource *LineupDetail `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	// An outsider (not on the team) cannot confirm.
	w = doJSON(t, router, http.MethodPost, "/api/lineup/"+created.Resource.Lineup.ID.String()+"/confirm", nil, newToken(t, fx.outsider))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider confirm: %s", w.Body.String())

	// A non-confirmed lineup cannot be marked official by the commissioner.
	w = doJSON(t, router, http.MethodPost, "/api/lineup/"+created.Resource.Lineup.ID.String()+"/official", nil, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusBadRequest, w.Code, "official before confirm: %s", w.Body.String())
}
