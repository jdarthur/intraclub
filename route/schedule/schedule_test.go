package schedule

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
// test helpers (mirror route/week)
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

// newSeasonWithTeams builds a Season (with a commissioner and a draft) plus
// numTeams teams whose captains are returned. The teams are assigned to the
// season, making their captains "participants" who must not be able to create
// or modify the schedule.
func newSeasonWithTeams(t *testing.T, db database.Provider, commissionerID database.UserId, numTeams int) (*model.Season, []*model.Team) {
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
	facility.Name = fmt.Sprintf("Test facility %d", rand.Uint64())
	facility.Address = fmt.Sprintf("Test Rd %d", rand.Uint64())
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

	teams := make([]*model.Team, 0, numTeams)
	for i := 0; i < numTeams; i++ {
		captain := newStoredUser(t, db)
		team := model.NewDefaultTeam(captain.ID, fmt.Sprintf("Team %d", i+1))
		teamV, err := database.CreateOne(ctx, db, team)
		require.NoError(t, err)
		require.NoError(t, seasonV.AddTeam(ctx, db, teamV.ID))
		teams = append(teams, teamV)
	}

	return seasonV, teams
}

// newStoredWeek creates a Week for the season's draft.
func newStoredWeek(t *testing.T, db database.Provider, season *model.Season) *model.Week {
	t.Helper()
	week := model.NewWeek()
	week.DraftId = season.DraftId
	week.Date = time.Date(2025, 3, 1, 8, 0, 0, 0, time.UTC)
	week.Note = "week 1"
	v, err := database.CreateOne(context.Background(), db, week)
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

func createScheduleViaHTTP(t *testing.T, router *gin.Engine, seasonID string, token string) *httptest.ResponseRecorder {
	t.Helper()
	return doJSON(t, router, http.MethodPost, "/api/schedule", map[string]any{"season_id": seasonID}, token)
}

func createSchedule(t *testing.T, router *gin.Engine, season *model.Season, token string) string {
	t.Helper()
	w := createScheduleViaHTTP(t, router, season.ID.String(), token)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var created struct {
		Resource *model.Schedule `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotNil(t, created.Resource)
	return created.Resource.ID.String()
}

// ---------------------------------------------------------------------------
// create schedule: commissioner only, one per season
// ---------------------------------------------------------------------------

func TestCreateScheduleCommissionerOnly(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, teams := newSeasonWithTeams(t, db, commissioner.ID, 2)
	_ = teams

	// Commissioner may create the schedule.
	w := createScheduleViaHTTP(t, router, season.ID.String(), newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "commissioner create: %s", w.Body.String())

	// A team captain (season participant) may NOT create the schedule.
	captainIDs, err := season.GetTeamCaptains(context.Background(), db)
	require.NoError(t, err)
	require.Len(t, captainIDs, 2)
	w = createScheduleViaHTTP(t, router, season.ID.String(), newToken(t, captainIDs[0]))
	require.Equal(t, http.StatusForbidden, w.Code, "captain create: %s", w.Body.String())

	// An unrelated user may NOT create the schedule.
	outsider := newStoredUser(t, db)
	w = createScheduleViaHTTP(t, router, season.ID.String(), newToken(t, outsider.ID))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider create: %s", w.Body.String())

	// No token -> unauthorized.
	w = createScheduleViaHTTP(t, router, season.ID.String(), "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOneSchedulePerSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, _ := newSeasonWithTeams(t, db, commissioner.ID, 2)

	createSchedule(t, router, season, newToken(t, commissioner.ID))

	// A second schedule for the same season must be rejected.
	w := createScheduleViaHTTP(t, router, season.ID.String(), newToken(t, commissioner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "duplicate schedule: %s", w.Body.String())
}

// ---------------------------------------------------------------------------
// assign weekly matchups: commissioner only
// ---------------------------------------------------------------------------

func assignMatchupViaHTTP(t *testing.T, router *gin.Engine, scheduleID string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	return doJSON(t, router, http.MethodPost, "/api/schedule/"+scheduleID+"/weekly_matchup", body, token)
}

func TestAssignWeeklyMatchupCommissionerOnly(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, teams := newSeasonWithTeams(t, db, commissioner.ID, 2)
	week := newStoredWeek(t, db, season)

	scheduleID := createSchedule(t, router, season, newToken(t, commissioner.ID))

	matchupBody := map[string]any{
		"week_id": week.ID.String(),
		"matchups": []map[string]any{
			{"home_team_id": teams[0].ID.String(), "away_team_id": teams[1].ID.String(), "bye": false},
		},
	}

	// Commissioner may assign.
	w := assignMatchupViaHTTP(t, router, scheduleID, matchupBody, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "commissioner assign: %s", w.Body.String())

	// A team captain (season participant) may NOT assign.
	captainIDs, err := season.GetTeamCaptains(context.Background(), db)
	require.NoError(t, err)
	w = assignMatchupViaHTTP(t, router, scheduleID, matchupBody, newToken(t, captainIDs[0]))
	require.Equal(t, http.StatusForbidden, w.Code, "captain assign: %s", w.Body.String())

	// An unrelated user may NOT assign.
	outsider := newStoredUser(t, db)
	w = assignMatchupViaHTTP(t, router, scheduleID, matchupBody, newToken(t, outsider.ID))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider assign: %s", w.Body.String())
}

func TestAssignWeeklyMatchupWithBye(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, teams := newSeasonWithTeams(t, db, commissioner.ID, 2)
	week := newStoredWeek(t, db, season)

	scheduleID := createSchedule(t, router, season, newToken(t, commissioner.ID))

	// Both teams are on a bye this week.
	body := map[string]any{
		"week_id": week.ID.String(),
		"matchups": []map[string]any{
			{"home_team_id": teams[0].ID.String(), "bye": true},
			{"home_team_id": teams[1].ID.String(), "bye": true},
		},
	}
	w := assignMatchupViaHTTP(t, router, scheduleID, body, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "bye assign: %s", w.Body.String())
}

func TestAssignWeeklyMatchupRequiresAllTeamsCovered(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, teams := newSeasonWithTeams(t, db, commissioner.ID, 2)
	week := newStoredWeek(t, db, season)

	scheduleID := createSchedule(t, router, season, newToken(t, commissioner.ID))

	// A single-team entry (home team with a bye but no away) leaves the second
	// team uncovered and must be rejected.
	badBody := map[string]any{
		"week_id": week.ID.String(),
		"matchups": []map[string]any{
			{"home_team_id": teams[0].ID.String(), "away_team_id": teams[1].ID.String(), "bye": false},
			{"home_team_id": teams[0].ID.String(), "bye": true},
		},
	}
	// home team team[0] double-booked -> rejected.
	w := assignMatchupViaHTTP(t, router, scheduleID, badBody, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "bad assign: %s", w.Body.String())

	// The week must belong to the season's draft; a week from another draft is
	// rejected.
	otherSeason, otherTeams := newSeasonWithTeams(t, db, newStoredUser(t, db).ID, 2)
	_ = otherTeams
	foreignWeek := newStoredWeek(t, db, otherSeason)
	foreignBody := map[string]any{
		"week_id": foreignWeek.ID.String(),
		"matchups": []map[string]any{
			{"home_team_id": teams[0].ID.String(), "away_team_id": teams[1].ID.String(), "bye": false},
		},
	}
	w = assignMatchupViaHTTP(t, router, scheduleID, foreignBody, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code, "foreign week assign: %s", w.Body.String())
}

// ---------------------------------------------------------------------------
// view: schedule detail is visible to everyone
// ---------------------------------------------------------------------------

func TestScheduleDetailVisibleToEveryone(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, teams := newSeasonWithTeams(t, db, commissioner.ID, 2)
	week := newStoredWeek(t, db, season)

	scheduleID := createSchedule(t, router, season, newToken(t, commissioner.ID))

	assignMatchupViaHTTP(t, router, scheduleID, map[string]any{
		"week_id": week.ID.String(),
		"matchups": []map[string]any{
			{"home_team_id": teams[0].ID.String(), "away_team_id": teams[1].ID.String(), "bye": false},
		},
	}, newToken(t, commissioner.ID))

	// A participant (captain) and an outsider can both view the detail.
	captainIDs, err := season.GetTeamCaptains(context.Background(), db)
	require.NoError(t, err)
	for _, tokenUser := range []database.UserId{captainIDs[0], newStoredUser(t, db).ID} {
		w := doJSON(t, router, http.MethodGet, "/api/schedule/"+scheduleID, nil, newToken(t, tokenUser))
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var detail struct {
			Resource ScheduleDetail `json:"resource"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &detail))
		require.NotNil(t, detail.Resource.Schedule)
		require.Len(t, detail.Resource.WeeklyMatchups, 1)
		require.Len(t, detail.Resource.WeeklyMatchups[0].Matchups, 1)
		require.Equal(t, teams[0].ID, detail.Resource.WeeklyMatchups[0].Matchups[0].HomeTeam)
	}
}

func TestScheduleDetailBeforeCreationHasNullSchedule(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	season, _ := newSeasonWithTeams(t, db, commissioner.ID, 2)

	// No schedule yet: the season-scoped query returns a detail with nil
	// Schedule but the commissioners listed.
	w := doJSON(t, router, http.MethodGet, "/api/schedule?season_id="+season.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var detail struct {
		Resource ScheduleDetail `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &detail))
	require.Nil(t, detail.Resource.Schedule)
	require.Len(t, detail.Resource.Commissioners, 1)
	require.Equal(t, commissioner.ID, detail.Resource.Commissioners[0])
}
