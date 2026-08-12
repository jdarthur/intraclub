package draft

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
// test setup helpers (mirror model/*_test.go helpers)
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

// newDefaultFormat creates a stored Format with two lines whose unique ratings
// become the format's possible ratings.
func newDefaultFormat(t *testing.T, db database.Provider) *model.Format {
	t.Helper()
	user := newStoredUser(t, db)
	f := model.NewFormat()
	f.UserId = user.ID
	f.Name = "default format"
	created, err := database.CreateOne(context.Background(), db, f)
	require.NoError(t, err)

	lines := []model.FormatLine{
		model.FormatLine{Player1Rating: newStoredRating(t, db).ID, Player2Rating: newStoredRating(t, db).ID},
		model.FormatLine{Player1Rating: newStoredRating(t, db).ID, Player2Rating: newStoredRating(t, db).ID},
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

func newFacility(t *testing.T, db database.Provider, owner database.UserId) *model.Facility {
	t.Helper()
	facility := model.NewFacility()
	facility.UserId = owner
	facility.Name = fmt.Sprintf("Test facility %s", database.NewRecordId())
	facility.Address = fmt.Sprintf("%s Test Rd.", database.NewRecordId())
	facility.NumberOfCourts = 5
	v, err := database.CreateOne(context.Background(), db, facility)
	require.NoError(t, err)
	return v
}

// newTestRouter builds a gin engine with the full draft REST surface registered
// (as in main.go) and the auth globals wired up so real tokens validate.
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

// createDraftViaHTTP creates a draft through the CRUD route as the owner.
func createDraftViaHTTP(t *testing.T, router *gin.Engine, owner database.UserId, formatID model.FormatId, name string) string {
	t.Helper()
	w := doJSON(t, router, http.MethodPost, "/api/draft", map[string]any{
		"name":   name,
		"format": formatID.String(),
	}, newToken(t, owner))
	require.Equal(t, http.StatusOK, w.Code, "create draft: %s", w.Body.String())

	var resp struct {
		Resource *model.Draft `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotNil(t, resp.Resource)
	return resp.Resource.ID.String()
}

// ---------------------------------------------------------------------------
// CRUD endpoints
// ---------------------------------------------------------------------------

func TestDraftCRUD(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)

	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "CRUD Draft")

	// GET one
	w := doJSON(t, router, http.MethodGet, "/api/draft/"+draftID, nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var one struct {
		Resource *model.Draft `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &one))
	require.Equal(t, draftID, one.Resource.ID.String())
	require.Equal(t, commissioner.ID, one.Resource.GetOwner())

	// GET many
	w = doJSON(t, router, http.MethodGet, "/api/draft", nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var many struct {
		Resource []*model.Draft `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &many))
	require.Len(t, many.Resource, 1)

	// PUT update
	w = doJSON(t, router, http.MethodPut, "/api/draft/"+draftID, map[string]any{
		"id":     draftID,
		"name":   "Renamed Draft",
		"format": format.ID.String(),
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "update draft: %s", w.Body.String())

	// DELETE
	w = doJSON(t, router, http.MethodDelete, "/api/draft/"+draftID, nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)

	w = doJSON(t, router, http.MethodGet, "/api/draft/"+draftID, nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestDraftJoinModelCRUD(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)
	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "Join CRUD")
	player := newStoredUser(t, db)

	// Create a DraftAvailablePlayer via CRUD.
	w := doJSON(t, router, http.MethodPost, "/api/draft_available_player", map[string]any{
		"draft_id":  draftID,
		"player_id": player.ID.String(),
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "create available player: %s", w.Body.String())
	var created struct {
		Resource *model.DraftAvailablePlayer `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotNil(t, created.Resource)

	// GET one
	w = doJSON(t, router, http.MethodGet, "/api/draft_available_player/"+created.Resource.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)

	// GET many
	w = doJSON(t, router, http.MethodGet, "/api/draft_available_player", nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)
	var many struct {
		Resource []*model.DraftAvailablePlayer `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &many))
	require.Len(t, many.Resource, 1)

	// DELETE
	w = doJSON(t, router, http.MethodDelete, "/api/draft_available_player/"+created.Resource.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// custom action endpoints
// ---------------------------------------------------------------------------

// mustParseRecordID converts a hex-string record ID back into a RecordId.
func mustParseRecordID(t *testing.T, s string) database.RecordId {
	t.Helper()
	id, err := database.RecordIdFromString(s)
	require.NoError(t, err)
	return id
}

// completeDraftViaHTTP drives the draft to completion through the select
// endpoint, authenticating each pick as the captain on the clock. It mirrors
// the model test helper's ordering: each captain first picks themselves (in
// draft order), then the remaining non-captain players are drafted.
func completeDraftViaHTTP(t *testing.T, router *gin.Engine, db database.Provider, draftID string, captainTokens map[database.UserId]string) {
	t.Helper()
	draftIDR := mustParseRecordID(t, draftID)

	// 1) each captain picks themselves, in draft order
	draft, err := database.GetExistingRecordById(context.Background(), db, &model.Draft{}, draftIDR)
	require.NoError(t, err)
	captains, err := draft.GetCaptains(context.Background(), db)
	require.NoError(t, err)
	for _, c := range captains {
		token, ok := captainTokens[c.CaptainId]
		require.True(t, ok, "no token for captain %s", c.CaptainId)
		w := doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/select", map[string]any{
			"player_id": c.CaptainId.String(),
		}, token)
		require.Equal(t, http.StatusOK, w.Code, "captain self-select: %s", w.Body.String())
	}

	// 2) draft the remaining players until the draft is complete
	for {
		draft, err = database.GetExistingRecordById(context.Background(), db, &model.Draft{}, draftIDR)
		require.NoError(t, err)
		if draft.IsDraftCompleted(context.Background(), db) {
			return
		}
		onTheClock, err := draft.GetCaptainOnTheClock(context.Background(), db)
		require.NoError(t, err)
		available := draft.GetAllAvailableToSelect(onTheClock, db)
		require.NotEmpty(t, available, "no players available to select")

		token, ok := captainTokens[onTheClock]
		require.True(t, ok, "no token for on-the-clock captain %s", onTheClock)
		w := doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/select", map[string]any{
			"player_id": available[0].String(),
		}, token)
		require.Equal(t, http.StatusOK, w.Code, "select pick: %s", w.Body.String())
	}
}

func TestDraftFullFlow(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)

	// Captains and extra players.
	captains := make([]*model.User, 4)
	captainTokens := make(map[database.UserId]string)
	for i := 0; i < 4; i++ {
		c := newStoredUser(t, db)
		captains[i] = c
		captainTokens[c.ID] = newToken(t, c.ID)
	}
	extraPlayers := []*model.User{newStoredUser(t, db), newStoredUser(t, db)}

	// Create the draft and select the draft order pattern by name.
	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "Full Flow Draft")
	w := doJSON(t, router, http.MethodPut, "/api/draft/"+draftID+"/draft_order_pattern", map[string]any{
		"draft_order_pattern": "Snake",
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "set pattern: %s", w.Body.String())

	// Initialize (creates teams + DraftCaptain + captain DraftAvailablePlayer rows).
	captainIDs := make([]string, len(captains))
	for i, c := range captains {
		captainIDs[i] = c.ID.String()
	}
	w = doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/initialize", map[string]any{
		"captains": captainIDs,
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "initialize: %s", w.Body.String())

	// Assign additional draftable players.
	extraIDs := make([]string, len(extraPlayers))
	for i, p := range extraPlayers {
		extraIDs[i] = p.ID.String()
	}
	w = doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/assign_draftable_players", map[string]any{
		"players": extraIDs,
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "assign draftable players: %s", w.Body.String())

	// Assign rating cutoffs for every rating except the lowest (the draft's
	// DynamicallyValid requires a complete, increasing set).
	possibleRatings, err := format.GetPossibleRatings(context.Background(), db)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(possibleRatings), 2)
	for i, r := range possibleRatings[:len(possibleRatings)-1] {
		w = doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/assign_rating_cutoff", map[string]any{
			"rating": r.String(),
			"cutoff": i + 1,
		}, newToken(t, commissioner.ID))
		require.Equal(t, http.StatusOK, w.Code, "assign rating cutoff: %s", w.Body.String())
	}

	// Drive the draft to completion via the select endpoint.
	completeDraftViaHTTP(t, router, db, draftID, captainTokens)

	// Verify the draft is now completed.
	draft, err := database.GetExistingRecordById(context.Background(), db, &model.Draft{}, mustParseRecordID(t, draftID))
	require.NoError(t, err)
	require.True(t, draft.IsDraftCompleted(context.Background(), db))

	// Assign drafted players to teams.
	w = doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/assign_drafted_players_to_teams", nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "assign drafted players: %s", w.Body.String())

	// Create the season.
	facility := newFacility(t, db, commissioner.ID)
	w = doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/create_season", map[string]any{
		"name":       "Season from draft",
		"facility":   facility.ID.String(),
		"start_time": "08:30",
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "create season: %s", w.Body.String())
	var seasonResp struct {
		Resource *model.Season `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &seasonResp))
	require.NotNil(t, seasonResp.Resource)
	require.Equal(t, "Season from draft", seasonResp.Resource.Name)

	// The join rows produced by the draft actions are readable via CRUD.
	for _, path := range []string{"/api/draft_captain", "/api/draft_pick", "/api/draft_rating_cutoff"} {
		w = doJSON(t, router, http.MethodGet, path, nil, newToken(t, commissioner.ID))
		require.Equal(t, http.StatusOK, w.Code, "GET %s: %s", path, w.Body.String())
		var list struct {
			Resource []json.RawMessage `json:"resource"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
		require.NotEmpty(t, list.Resource, "expected rows from %s", path)
	}
}

func TestDraftResultsEndpoint(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)

	captains := make([]*model.User, 3)
	captainTokens := make(map[database.UserId]string)
	for i := 0; i < 3; i++ {
		c := newStoredUser(t, db)
		captains[i] = c
		captainTokens[c.ID] = newToken(t, c.ID)
	}

	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "Results Draft")
	captainIDs := make([]string, len(captains))
	for i, c := range captains {
		captainIDs[i] = c.ID.String()
	}
	w := doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/initialize", map[string]any{
		"captains": captainIDs,
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "initialize: %s", w.Body.String())

	// Assign rating cutoffs for every rating except the lowest.
	draft, err := database.GetExistingRecordById(context.Background(), db, &model.Draft{}, mustParseRecordID(t, draftID))
	require.NoError(t, err)
	possibleRatings, err := format.GetPossibleRatings(context.Background(), db)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(possibleRatings), 2)
	for i, r := range possibleRatings[:len(possibleRatings)-1] {
		w = doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/assign_rating_cutoff", map[string]any{
			"rating": r.String(),
			"cutoff": i + 1,
		}, newToken(t, commissioner.ID))
		require.Equal(t, http.StatusOK, w.Code, "assign rating cutoff: %s", w.Body.String())
	}

	// Drive the draft to completion so every team has selections.
	completeDraftViaHTTP(t, router, db, draftID, captainTokens)
	require.True(t, draft.IsDraftCompleted(context.Background(), db))

	// The results endpoint returns each team (in draft order) with its roster
	// and per-player assigned ratings. It is readable by a non-owner user.
	outsider := newStoredUser(t, db)
	w = doJSON(t, router, http.MethodGet, "/api/draft/"+draftID+"/results", nil, newToken(t, outsider.ID))
	require.Equal(t, http.StatusOK, w.Code, "results: %s", w.Body.String())

	var resp struct {
		Resource struct {
			Teams []struct {
				TeamId     string `json:"team_id"`
				CaptainId  string `json:"captain_id"`
				DraftOrder int    `json:"draft_order"`
				Selections []struct {
					Round  int    `json:"round"`
					Pick   int    `json:"pick"`
					Rating string `json:"rating"`
					User   struct {
						ID string `json:"id"`
					} `json:"user"`
				} `json:"selections"`
			} `json:"teams"`
		} `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Resource.Teams, len(captains))

	// Teams are returned in draft order and every selection has a rating.
	selectionsTotal := 0
	for i, team := range resp.Resource.Teams {
		require.Equal(t, i, team.DraftOrder)
		require.Equal(t, captains[i].ID.String(), team.CaptainId)
		require.NotEmpty(t, team.TeamId)
		for _, sel := range team.Selections {
			require.Greater(t, sel.Round, 0)
			require.Greater(t, sel.Pick, 0)
			require.NotEmpty(t, sel.Rating)
			require.NotEmpty(t, sel.User.ID)
		}
		selectionsTotal += len(team.Selections)
	}
	// Every drafted player (captains + extra players) is accounted for.
	require.Equal(t, len(captains), selectionsTotal)
}

func TestDraftRatingCutoffCRUD(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)
	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "Cutoff CRUD")

	rating := newStoredRating(t, db)
	w := doJSON(t, router, http.MethodPost, "/api/draft_rating_cutoff", map[string]any{
		"draft_id":     draftID,
		"rating_id":    rating.ID.String(),
		"cutoff_index": 2,
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "create cutoff: %s", w.Body.String())
	var created struct {
		Resource *model.DraftRatingCutoff `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotNil(t, created.Resource)

	// GET one
	w = doJSON(t, router, http.MethodGet, "/api/draft_rating_cutoff/"+created.Resource.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)

	// DELETE
	w = doJSON(t, router, http.MethodDelete, "/api/draft_rating_cutoff/"+created.Resource.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestPreDraftGradeCRUD(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)
	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "Grade CRUD")

	// The graded player must be on the draft's available list.
	player := newStoredUser(t, db)
	_, err := database.CreateOne(context.Background(), db, &model.DraftAvailablePlayer{
		DraftId:  model.DraftId(mustParseRecordID(t, draftID)),
		PlayerId: player.ID,
	})
	require.NoError(t, err)

	possibleRatings, err := format.GetPossibleRatings(context.Background(), db)
	require.NoError(t, err)
	require.NotEmpty(t, possibleRatings)

	w := doJSON(t, router, http.MethodPost, "/api/pre_draft_grade", map[string]any{
		"PlayerId": player.ID.String(),
		"DraftId":  draftID,
		"Modifier": 0,
		"Rating":   possibleRatings[0].String(),
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code, "create grade: %s", w.Body.String())
	var created struct {
		Resource *model.PreDraftGrade `json:"resource"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotNil(t, created.Resource)
	// The grader is set from the token on creation.
	require.Equal(t, commissioner.ID, created.Resource.GraderId)

	// GET one
	w = doJSON(t, router, http.MethodGet, "/api/pre_draft_grade/"+created.Resource.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)

	// DELETE
	w = doJSON(t, router, http.MethodDelete, "/api/pre_draft_grade/"+created.Resource.ID.String(), nil, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestDraftInitializeAuthorization(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)
	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "Auth Draft")

	// A non-owner user may not initialize the draft.
	outsider := newStoredUser(t, db)
	captain := newStoredUser(t, db)
	w := doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/initialize", map[string]any{
		"captains": []string{captain.ID.String()},
	}, newToken(t, outsider.ID))
	require.Equal(t, http.StatusForbidden, w.Code)
}

// TestDraftInitializeEmptyCaptains verifies that body validation is enforced
// over the wire: an initialize request with no captains is rejected.
func TestDraftInitializeEmptyCaptains(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)
	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "Empty Captains")

	w := doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/initialize", map[string]any{
		"captains": []string{},
	}, newToken(t, commissioner.ID))
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDraftSelectAuthorization(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	router := newTestRouter(t, db)
	commissioner := newStoredUser(t, db)
	format := newDefaultFormat(t, db)
	draftID := createDraftViaHTTP(t, router, commissioner.ID, format.ID, "Select Auth Draft")

	// A non-captain may not make a selection.
	outsider := newStoredUser(t, db)
	player := newStoredUser(t, db)
	w := doJSON(t, router, http.MethodPost, "/api/draft/"+draftID+"/select", map[string]any{
		"player_id": player.ID.String(),
	}, newToken(t, outsider.ID))
	require.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------------------------------------------------------------------
// body validation
// ---------------------------------------------------------------------------

func TestInitializeBodyValidation(t *testing.T) {
	b := &InitializeBody{}
	require.Error(t, b.StaticallyValid())
	b.Captains = []database.UserId{database.UserId(database.NewRecordId())}
	require.NoError(t, b.StaticallyValid())
}

func TestSelectBodyValidation(t *testing.T) {
	b := &SelectBody{}
	require.Error(t, b.StaticallyValid())
	b.PlayerId = database.UserId(database.NewRecordId())
	require.NoError(t, b.StaticallyValid())
}

func TestAssignRatingCutoffBodyValidation(t *testing.T) {
	b := &AssignRatingCutoffBody{Cutoff: 0}
	require.Error(t, b.StaticallyValid())
	b.Cutoff = 1
	require.NoError(t, b.StaticallyValid())
}

func TestSetDraftOrderPatternBodyValidation(t *testing.T) {
	b := &SetDraftOrderPatternBody{}
	require.Error(t, b.StaticallyValid())
	b.DraftOrderPattern = "Snake"
	require.NoError(t, b.StaticallyValid())
}

func TestCreateSeasonBodyValidation(t *testing.T) {
	b := &CreateSeasonBody{}
	require.Error(t, b.StaticallyValid())
	b.Name = "season"
	require.NoError(t, b.StaticallyValid())
}
