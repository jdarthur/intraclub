package match

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
	"intraclub/route/week"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// test setup helpers (mirror route/lineup)
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

// newScoringStructure builds a simple set-scoring structure (win at 6 games,
// by 2) that Victorious can decide on.
func newScoringStructure(t *testing.T, db database.Provider, owner database.UserId) *model.ScoringStructure {
	t.Helper()
	s := model.NewScoringStructure()
	s.Owner = owner
	s.Name = fmt.Sprintf("scoring %d", rand.Uint64())
	s.WinConditionCountingType = model.Game
	s.WinCondition = model.WinCondition{WinThreshold: 6, MustWinBy: 2, InstantWinThreshold: 7}
	v, err := database.CreateOne(context.Background(), db, s)
	require.NoError(t, err)
	return v
}

// newRatedTeam creates a team with a captain (rating1) and a member (rating2)
// and assigns it to the season, so a format-line pairing (rating1/rating2) can
// be built for it.
func newRatedTeam(t *testing.T, db database.Provider, season *model.Season, rating1, rating2 model.RatingId, name string) (*model.Team, database.UserId, database.UserId) {
	t.Helper()
	ctx := context.Background()
	captain := newStoredUser(t, db)
	member := newStoredUser(t, db)
	team := model.NewDefaultTeam(captain.ID, name)
	teamV, err := database.CreateOne(ctx, db, team)
	require.NoError(t, err)
	require.NoError(t, season.AddTeam(ctx, db, teamV.ID))
	addAssignment(t, db, teamV, member, model.TeamRoleMember)
	_, err = database.CreateOne(ctx, db, &model.TeamRating{TeamId: teamV.ID, UserId: captain.ID, RatingId: rating1})
	require.NoError(t, err)
	_, err = database.CreateOne(ctx, db, &model.TeamRating{TeamId: teamV.ID, UserId: member.ID, RatingId: rating2})
	require.NoError(t, err)
	return teamV, captain.ID, member.ID
}

// matchFixture builds the full chain required to generate matches: a season
// with a schedule and a weekly matchup between two rated teams, a week, a
// scoring structure, and official lineups for both teams.
type matchFixture struct {
	season       *model.Season
	week         *model.Week
	homeTeam     *model.Team
	awayTeam     *model.Team
	homeCaptain  database.UserId
	awayCaptain  database.UserId
	scoring      *model.ScoringStructure
	commissioner database.UserId
	outsider     database.UserId
}

func newMatchFixture(t *testing.T, db database.Provider) *matchFixture {
	t.Helper()
	ctx := context.Background()
	commissioner := newStoredUser(t, db)
	outsider := newStoredUser(t, db)

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

	homeTeam, homeCaptain, _ := newRatedTeam(t, db, seasonV, rating1, rating2, "Home Team")
	awayTeam, awayCaptain, _ := newRatedTeam(t, db, seasonV, rating1, rating2, "Away Team")

	week := model.NewWeek()
	week.DraftId = draftV.ID
	week.Date = time.Date(2025, 3, 1, 8, 0, 0, 0, time.UTC)
	weekV, err := database.CreateOne(ctx, db, week)
	require.NoError(t, err)

	scoring := newScoringStructure(t, db, commissioner.ID)

	schedule := model.NewSchedule()
	schedule.SeasonId = seasonV.ID
	scheduleV, err := database.CreateOne(ctx, db, schedule)
	require.NoError(t, err)

	wm := model.NewWeeklyMatchup()
	wm.WeekId = weekV.ID
	wm.SeasonId = seasonV.ID
	wmV, err := database.CreateOne(ctx, db, wm)
	require.NoError(t, err)

	_, err = database.CreateOne(ctx, db, &model.ScheduleMatchup{ScheduleId: scheduleV.ID, WeeklyMatchupId: wmV.ID, Position: 0})
	require.NoError(t, err)
	_, err = database.CreateOne(ctx, db, model.NewWeeklyMatchupTeamMatchup(wmV.ID, homeTeam.ID, awayTeam.ID, false, 0))
	require.NoError(t, err)

	// Build and mark official a one-line lineup for each team.
	buildOfficialLineup(t, db, homeTeam.ID, weekV.ID, homeCaptain, rating1, rating2)
	buildOfficialLineup(t, db, awayTeam.ID, weekV.ID, awayCaptain, rating1, rating2)

	return &matchFixture{
		season:       seasonV,
		week:         weekV,
		homeTeam:     homeTeam,
		awayTeam:     awayTeam,
		homeCaptain:  homeCaptain,
		awayCaptain:  awayCaptain,
		scoring:      scoring,
		commissioner: commissioner.ID,
		outsider:     outsider.ID,
	}
}

// buildOfficialLineup creates a lineup (confirmed + official) with a single
// pairing of the team's captain/member at format line index 0.
func buildOfficialLineup(t *testing.T, db database.Provider, teamID model.TeamId, weekID model.WeekId, captain database.UserId, rating1, rating2 model.RatingId) {
	t.Helper()
	ctx := context.Background()
	team, err := database.GetExistingRecordById(ctx, db, &model.Team{}, teamID.RecordId())
	require.NoError(t, err)
	var memberID database.UserId
	assignments, err := database.GetAllWhere[*model.TeamAssignment](ctx, db, func(_ context.Context, a *model.TeamAssignment) bool {
		return a.TeamId == teamID && a.Role == model.TeamRoleMember
	})
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	memberID = assignments[0].UserId

	lineup := &model.Lineup{TeamId: teamID, WeekId: weekID, Confirmed: true, Official: true}
	lineupV, err := database.CreateOne(ctx, db, lineup)
	require.NoError(t, err)

	_, err = database.CreateOne(ctx, db, &model.LineupPairing{
		LineupId:        lineupV.ID,
		TeamId:          teamID,
		Player1:         captain,
		Player2:         memberID,
		FormatLineIndex: 0,
	})
	require.NoError(t, err)
	_ = team
	_ = rating1
	_ = rating2
}

func newTestRouter(t *testing.T, db database.Provider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	api.UserType = &model.User{}
	database.SysAdminCheck = model.IsUserSystemAdministrator

	router := gin.New()
	group := router.Group("/api")
	RegisterRoutes(group, db)
	week.RegisterRoutes(group, db)
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

// weekMatchDetail mirrors the wire shape returned by GET /match/week.
type weekMatchDetail struct {
	WeekId      string          `json:"week_id"`
	SeasonId    string          `json:"season_id"`
	Closed      bool            `json:"closed"`
	TeamMatches []*teamMatchDTO `json:"team_matches"`
}

type teamMatchDTO struct {
	ID       string                `json:"id"`
	HomeTeam string                `json:"home_team_id"`
	AwayTeam string                `json:"away_team_id"`
	Complete bool                  `json:"complete"`
	HomeWins int                   `json:"home_wins"`
	AwayWins int                   `json:"away_wins"`
	Winner   string                `json:"winner_team_id"`
	Matches  []*individualMatchDTO `json:"matches"`
}

type individualMatchDTO struct {
	ID       string `json:"id"`
	TeamId   string `json:"team_id"`
	Status   int    `json:"status"`
	Main     int    `json:"main_value"`
	Opponent string `json:"opponent"`
}

func generateMatches(t *testing.T, router *gin.Engine, fx *matchFixture, token string) *httptest.ResponseRecorder {
	t.Helper()
	return doJSON(t, router, http.MethodPost, "/api/match/generate", map[string]any{
		"week_id":              fx.week.ID.String(),
		"scoring_structure_id": fx.scoring.ID.String(),
	}, token)
}

func getWeekDetail(t *testing.T, router *gin.Engine, fx *matchFixture) *weekMatchDetail {
	t.Helper()
	w := doJSON(t, router, http.MethodGet, "/api/match/week?week_id="+fx.week.ID.String(), nil, "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var body struct {
		Resource *weekMatchDetail `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.NotNil(t, body.Resource)
	return body.Resource
}

// ---------------------------------------------------------------------------
// tests
// ---------------------------------------------------------------------------

func TestGenerateMatchesCommissionerOnly(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newMatchFixture(t, db)

	// No token -> unauthorized.
	w := generateMatches(t, router, fx, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())

	// An unrelated user may NOT generate matches.
	w = generateMatches(t, router, fx, newToken(t, fx.outsider))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider generate: %s", w.Body.String())

	// The commissioner generates the week's matches.
	w = generateMatches(t, router, fx, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, "commissioner generate: %s", w.Body.String())

	// The week score sheet now has one team match with two individual matches
	// (one per side).
	detail := getWeekDetail(t, router, fx)
	require.Len(t, detail.TeamMatches, 1)
	tm := detail.TeamMatches[0]
	require.Equal(t, fx.homeTeam.ID.String(), tm.HomeTeam)
	require.Equal(t, fx.awayTeam.ID.String(), tm.AwayTeam)
	require.Len(t, tm.Matches, 2)
	require.False(t, tm.Complete)
}

func TestRecordScoreAndComplete(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newMatchFixture(t, db)

	w := generateMatches(t, router, fx, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	detail := getWeekDetail(t, router, fx)
	tm := detail.TeamMatches[0]
	require.Len(t, tm.Matches, 2)

	// Identify the home and away sides.
	var homeMatch, awayMatch *individualMatchDTO
	for _, m := range tm.Matches {
		if m.TeamId == fx.homeTeam.ID.String() {
			homeMatch = m
		} else {
			awayMatch = m
		}
	}
	require.NotNil(t, homeMatch)
	require.NotNil(t, awayMatch)

	// An unrelated user cannot record a score.
	w = doJSON(t, router, http.MethodPost, "/api/match/score", map[string]any{
		"individual_match_id": homeMatch.ID,
		"main_value":          6,
	}, newToken(t, fx.outsider))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider score: %s", w.Body.String())

	// The commissioner records 6-3 for the home side.
	w = doJSON(t, router, http.MethodPost, "/api/match/score", map[string]any{
		"individual_match_id": homeMatch.ID,
		"main_value":          6,
		"secondary_value":     3,
	}, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, "home score: %s", w.Body.String())

	// Completing before the away side is scored must fail.
	w = doJSON(t, router, http.MethodPost, "/api/match/"+homeMatch.ID+"/complete", nil, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusBadRequest, w.Code, "complete early: %s", w.Body.String())

	// Record the away score (3-6) then complete the home side -> home wins.
	w = doJSON(t, router, http.MethodPost, "/api/match/score", map[string]any{
		"individual_match_id": awayMatch.ID,
		"main_value":          3,
		"secondary_value":     6,
	}, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, "away score: %s", w.Body.String())

	w = doJSON(t, router, http.MethodPost, "/api/match/"+homeMatch.ID+"/complete", nil, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, "complete: %s", w.Body.String())

	// The team match is now complete with a home win.
	detail = getWeekDetail(t, router, fx)
	tm = detail.TeamMatches[0]
	require.True(t, tm.Complete)
	require.Equal(t, 1, tm.HomeWins)
	require.Equal(t, 0, tm.AwayWins)
	require.Equal(t, fx.homeTeam.ID.String(), tm.Winner)
}

func TestCloseWeek(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newMatchFixture(t, db)

	w := generateMatches(t, router, fx, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	detail := getWeekDetail(t, router, fx)
	tm := detail.TeamMatches[0]

	// A week with an incomplete match cannot be closed.
	w = doJSON(t, router, http.MethodPost, "/api/week/"+fx.week.ID.String()+"/close", nil, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusBadRequest, w.Code, "close incomplete: %s", w.Body.String())

	// Score both sides and complete.
	homeMatch, awayMatch := tm.Matches[0], tm.Matches[1]
	score := func(id string, main int) {
		w := doJSON(t, router, http.MethodPost, "/api/match/score", map[string]any{
			"individual_match_id": id,
			"main_value":          main,
		}, newToken(t, fx.commissioner))
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	}
	score(homeMatch.ID, 6)
	score(awayMatch.ID, 3)
	w = doJSON(t, router, http.MethodPost, "/api/match/"+homeMatch.ID+"/complete", nil, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// An unrelated user cannot close the week.
	w = doJSON(t, router, http.MethodPost, "/api/week/"+fx.week.ID.String()+"/close", nil, newToken(t, fx.outsider))
	require.Equal(t, http.StatusForbidden, w.Code, "outsider close: %s", w.Body.String())

	// The commissioner closes the week.
	w = doJSON(t, router, http.MethodPost, "/api/week/"+fx.week.ID.String()+"/close", nil, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, "close: %s", w.Body.String())

	detail = getWeekDetail(t, router, fx)
	require.True(t, detail.Closed)
}

func TestStandings(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	fx := newMatchFixture(t, db)

	w := generateMatches(t, router, fx, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	detail := getWeekDetail(t, router, fx)
	tm := detail.TeamMatches[0]
	homeMatch, awayMatch := tm.Matches[0], tm.Matches[1]

	// Home wins 6-3.
	doJSON(t, router, http.MethodPost, "/api/match/score", map[string]any{"individual_match_id": homeMatch.ID, "main_value": 6}, newToken(t, fx.commissioner))
	doJSON(t, router, http.MethodPost, "/api/match/score", map[string]any{"individual_match_id": awayMatch.ID, "main_value": 3}, newToken(t, fx.commissioner))
	w = doJSON(t, router, http.MethodPost, "/api/match/"+homeMatch.ID+"/complete", nil, newToken(t, fx.commissioner))
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// Standings reflect the home team's win.
	w = doJSON(t, router, http.MethodGet, "/api/match/standings?season_id="+fx.season.ID.String(), nil, "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var body struct {
		Resource []*standingsEntry `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Resource, 2)
	homeEntry, awayEntry := body.Resource[0], body.Resource[1]
	if homeEntry.TeamId != fx.homeTeam.ID.String() {
		homeEntry, awayEntry = awayEntry, homeEntry
	}
	require.Equal(t, fx.homeTeam.ID.String(), homeEntry.TeamId)
	require.Equal(t, 1, homeEntry.Wins)
	require.Equal(t, 0, homeEntry.Losses)
	require.Equal(t, fx.awayTeam.ID.String(), awayEntry.TeamId)
	require.Equal(t, 1, awayEntry.Losses)
}

type standingsEntry struct {
	TeamId string `json:"team_id"`
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Ties   int    `json:"ties"`
}
