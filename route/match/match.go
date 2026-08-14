package match

import (
	"context"
	"errors"
	"net/http"
	"sort"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// BaseRoute is the base path for the Match REST surface.
const BaseRoute = "/match"

// isSeasonCommissioner reports whether the requesting user is one of the
// season's commissioners (the only role allowed to generate matches and close
// a week).
func isSeasonCommissioner(ctx context.Context, db database.Provider, userId database.UserId, seasonId model.SeasonId) bool {
	if userId == database.InvalidUserId {
		return false
	}
	for _, uid := range model.EditableBySeason(ctx, db, seasonId) {
		if uid == userId {
			return true
		}
	}
	return false
}

// canEditMatch reports whether the requesting user may record/complete a score
// on an individual match: they are a listed match editor (the season
// commissioners are added as editors when matches are generated), or a
// sysadmin.
func canEditMatch(ctx context.Context, db database.Provider, userId database.UserId, matchId model.IndividualMatchId) bool {
	if userId == database.InvalidUserId {
		return false
	}
	editors, err := database.GetAllWhere[*model.MatchEditor](ctx, db, func(_ context.Context, e *model.MatchEditor) bool {
		return e.MatchId == matchId && e.EditorUserId == userId
	})
	if err == nil && len(editors) > 0 {
		return true
	}
	if isAdmin, err := model.IsUserSystemAdministrator(ctx, db, userId); err == nil && isAdmin {
		return true
	}
	return false
}

// weekSeason resolves the season a week belongs to via its draft.
func weekSeason(ctx context.Context, db database.Provider, week *model.Week) (*model.Season, error) {
	draft, err := database.GetExistingRecordById(ctx, db, &model.Draft{}, week.DraftId.RecordId())
	if err != nil {
		return nil, err
	}
	return draft.GetSeason(ctx, db)
}

// officialLineupForTeamWeek returns the team's official lineup for the week, or
// nil if the team has not yet had one marked official.
func officialLineupForTeamWeek(ctx context.Context, db database.Provider, teamId model.TeamId, weekId model.WeekId) (*model.Lineup, error) {
	lineups, err := database.GetAllWhere[*model.Lineup](ctx, db, func(_ context.Context, l *model.Lineup) bool {
		return l.TeamId == teamId && l.WeekId == weekId && l.Official
	})
	if err != nil {
		return nil, err
	}
	if len(lineups) == 0 {
		return nil, nil
	}
	return lineups[0], nil
}

// pairingsForLineup returns a lineup's pairings sorted by format line index.
func pairingsForLineup(ctx context.Context, db database.Provider, lineupId model.LineupId) ([]*model.LineupPairing, error) {
	pairings, err := database.GetAllWhere[*model.LineupPairing](ctx, db, func(_ context.Context, p *model.LineupPairing) bool {
		return p.LineupId == lineupId
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(pairings, func(i, j int) bool {
		return pairings[i].FormatLineIndex < pairings[j].FormatLineIndex
	})
	return pairings, nil
}

// hasScore reports whether a side of an individual match has been scored yet.
func hasScore(m *model.IndividualMatch) bool {
	return m.WinOverride || m.MainValue > 0 || m.SecondaryValue > 0
}

// IndividualMatchDTO is the wire representation of one scored line in a team
// match, including which team/pairing it belongs to and the opponent's status.
type IndividualMatchDTO struct {
	ID              string `json:"id"`
	Structure       string `json:"structure"`
	MainValue       int    `json:"main_value"`
	SecondaryValue  int    `json:"secondary_value"`
	WinOverride     bool   `json:"win_override"`
	Status          int    `json:"status"`
	TeamId          string `json:"team_id"`
	Player1         string `json:"player1"`
	Player2         string `json:"player2"`
	FormatLineIndex int    `json:"format_line_index"`
	Opponent        string `json:"opponent"`
	OpponentStatus  int    `json:"opponent_status"`
}

// TeamMatchDTO is the wire representation of one head-to-head team match with
// its individual matches and a running win tally.
type TeamMatchDTO struct {
	ID       string              `json:"id"`
	WeekId   string              `json:"week_id"`
	HomeTeam string              `json:"home_team_id"`
	AwayTeam string              `json:"away_team_id"`
	Complete bool                `json:"complete"`
	HomeWins int                 `json:"home_wins"`
	AwayWins int                 `json:"away_wins"`
	Winner   string              `json:"winner_team_id"`
	Matches  []*IndividualMatchDTO `json:"matches"`
}

// WeekMatchDetail is the wire representation of a week's score sheet.
type WeekMatchDetail struct {
	WeekId      string          `json:"week_id"`
	SeasonId    string          `json:"season_id"`
	Closed      bool            `json:"closed"`
	TeamMatches []*TeamMatchDTO `json:"team_matches"`
}

// buildIndividualMatchDTO converts a stored individual match plus its pairing
// into its wire form.
func buildIndividualMatchDTO(ctx context.Context, db database.Provider, im *model.IndividualMatch, pairing *model.LineupPairing) (*IndividualMatchDTO, error) {
	dto := &IndividualMatchDTO{
		ID:              im.ID.RecordId().String(),
		Structure:       im.Structure.RecordId().String(),
		MainValue:       im.MainValue,
		SecondaryValue:  im.SecondaryValue,
		WinOverride:     im.WinOverride,
		Status:          int(im.Status),
		TeamId:          pairing.TeamId.RecordId().String(),
		Player1:         database.RecordId(pairing.Player1).String(),
		Player2:         database.RecordId(pairing.Player2).String(),
		FormatLineIndex: pairing.FormatLineIndex,
		Opponent:        im.Opponent.RecordId().String(),
	}
	if im.Opponent != model.IndividualMatchId(database.InvalidRecordId) {
		opp, err := database.GetExistingRecordById(ctx, db, &model.IndividualMatch{}, im.Opponent.RecordId())
		if err != nil {
			return nil, err
		}
		dto.OpponentStatus = int(opp.Status)
	}
	return dto, nil
}

// buildTeamMatchDTO assembles a TeamMatch with its individual matches and win
// tally from the join table and each assigned lineup pairing.
func buildTeamMatchDTO(ctx context.Context, db database.Provider, tm *model.TeamMatch) (*TeamMatchDTO, error) {
	rows, err := database.GetAllWhere[*model.TeamMatchIndividualMatch](ctx, db, func(_ context.Context, r *model.TeamMatchIndividualMatch) bool {
		return r.TeamMatchId == tm.ID
	})
	if err != nil {
		return nil, err
	}
	dto := &TeamMatchDTO{
		ID:       tm.ID.RecordId().String(),
		WeekId:   tm.WeekId.RecordId().String(),
		HomeTeam: tm.HomeTeam.RecordId().String(),
		AwayTeam: tm.AwayTeam.RecordId().String(),
		Matches:  []*IndividualMatchDTO{},
	}
	complete := true
	for _, row := range rows {
		pairing, err := database.GetExistingRecordById(ctx, db, &model.LineupPairing{}, row.LineupPairingId.RecordId())
		if err != nil {
			return nil, err
		}
		im, err := database.GetExistingRecordById(ctx, db, &model.IndividualMatch{}, row.IndividualMatchId.RecordId())
		if err != nil {
			return nil, err
		}
		matchDTO, err := buildIndividualMatchDTO(ctx, db, im, pairing)
		if err != nil {
			return nil, err
		}
		dto.Matches = append(dto.Matches, matchDTO)
		if im.Status == model.MatchWon {
			if pairing.TeamId == tm.HomeTeam {
				dto.HomeWins++
			} else if pairing.TeamId == tm.AwayTeam {
				dto.AwayWins++
			}
		}
		if im.Status != model.MatchWon && im.Status != model.MatchLost {
			complete = false
		}
	}
	dto.Complete = complete
	if complete && dto.HomeWins > dto.AwayWins {
		dto.Winner = tm.HomeTeam.RecordId().String()
	} else if complete && dto.AwayWins > dto.HomeWins {
		dto.Winner = tm.AwayTeam.RecordId().String()
	}
	return dto, nil
}

// weekDetailForWeek assembles the WeekMatchDetail for a week.
func weekDetailForWeek(ctx context.Context, db database.Provider, week *model.Week) (*WeekMatchDetail, error) {
	season, err := weekSeason(ctx, db, week)
	if err != nil {
		return nil, err
	}
	detail := &WeekMatchDetail{WeekId: week.ID.RecordId().String(), SeasonId: season.ID.RecordId().String(), Closed: week.Closed, TeamMatches: []*TeamMatchDTO{}}
	teamMatches, err := database.GetAllWhere[*model.TeamMatch](ctx, db, func(_ context.Context, tm *model.TeamMatch) bool {
		return tm.WeekId == week.ID
	})
	if err != nil {
		return nil, err
	}
	for _, tm := range teamMatches {
		dto, err := buildTeamMatchDTO(ctx, db, tm)
		if err != nil {
			return nil, err
		}
		detail.TeamMatches = append(detail.TeamMatches, dto)
	}
	return detail, nil
}

// ---- request/response types ----

// EmptyBody is a placeholder for routes that carry no request body.
type EmptyBody struct{}

// StaticallyValid has no static constraints.
func (b *EmptyBody) StaticallyValid() error { return nil }

// GenerateBody is the request body for GenerateMatches.
type GenerateBody struct {
	WeekId             string `json:"week_id"`
	ScoringStructureId string `json:"scoring_structure_id"`
}

// StaticallyValid ensures a week and a scoring structure are specified.
func (b *GenerateBody) StaticallyValid() error {
	if b.WeekId == "" {
		return errors.New("week_id must be set")
	}
	if b.ScoringStructureId == "" {
		return errors.New("scoring_structure_id must be set")
	}
	return nil
}

// GenerateMatches creates the week's TeamMatches (and their individual matches)
// from the scheduled weekly matchups and both teams' official lineups. Only a
// season commissioner may generate matches.
type GenerateMatches struct{}

func (c GenerateMatches) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute + "/generate"
}

func (c GenerateMatches) RequestBody() (*GenerateBody, bool) {
	return &GenerateBody{}, true
}

func (c GenerateMatches) Handler(req api.Request[*GenerateBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}
	weekRid, err := database.RecordIdFromString(req.Body.WeekId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	ssRid, err := database.RecordIdFromString(req.Body.ScoringStructureId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	week, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Week{}, weekRid)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	season, err := weekSeason(req.Context, req.DatabaseProvider, week)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if !isSeasonCommissioner(req.Context, req.DatabaseProvider, req.Token.UserId, season.ID) {
		return nil, http.StatusForbidden, errors.New("only a season commissioner may generate matches")
	}
	if _, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.ScoringStructure{}, ssRid); err != nil {
		return nil, http.StatusBadRequest, err
	}
	detail, err := generateWeekMatches(req.Context, req.DatabaseProvider, week, season, model.ScoringStructureId(ssRid))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: detail}, http.StatusOK, nil
}

// generateWeekMatches creates team + individual matches for a week from the
// scheduled weekly matchup entries and both teams' official lineups, pairing
// home and away lineup slots by format line index.
func generateWeekMatches(ctx context.Context, db database.Provider, week *model.Week, season *model.Season, scoringStructureId model.ScoringStructureId) (*WeekMatchDetail, error) {
	if season.ScheduleID.RecordId() == database.InvalidRecordId {
		return nil, errors.New("season has no schedule")
	}
	schedule, err := database.GetExistingRecordById(ctx, db, &model.Schedule{}, season.ScheduleID.RecordId())
	if err != nil {
		return nil, err
	}
	weeklyMatchups, err := schedule.GetMatchups(ctx, db)
	if err != nil {
		return nil, err
	}
	var wm *model.WeeklyMatchup
	for _, m := range weeklyMatchups {
		if m.WeekId == week.ID {
			wm = m
			break
		}
	}
	if wm == nil {
		return nil, errors.New("week has no assigned weekly matchup")
	}
	entries, err := wm.GetMatchups(ctx, db)
	if err != nil {
		return nil, err
	}
	commissioners, err := season.GetCommissioners(ctx, db)
	if err != nil {
		return nil, err
	}

	detail := &WeekMatchDetail{WeekId: week.ID.RecordId().String(), SeasonId: season.ID.RecordId().String(), Closed: week.Closed, TeamMatches: []*TeamMatchDTO{}}
	for _, entry := range entries {
		if entry.Bye {
			continue
		}
		homeLineup, err := officialLineupForTeamWeek(ctx, db, entry.HomeTeam, week.ID)
		if err != nil {
			return nil, err
		}
		awayLineup, err := officialLineupForTeamWeek(ctx, db, entry.AwayTeam, week.ID)
		if err != nil {
			return nil, err
		}
		if homeLineup == nil {
			return nil, errors.New("home team has no official lineup for this week")
		}
		if awayLineup == nil {
			return nil, errors.New("away team has no official lineup for this week")
		}

		tm := &model.TeamMatch{WeekId: week.ID, HomeTeam: entry.HomeTeam, AwayTeam: entry.AwayTeam, Lineup: homeLineup.ID}
		tmCreated, err := database.CreateOne(ctx, db, tm)
		if err != nil {
			return nil, err
		}

		homePairings, err := pairingsForLineup(ctx, db, homeLineup.ID)
		if err != nil {
			return nil, err
		}
		awayPairings, err := pairingsForLineup(ctx, db, awayLineup.ID)
		if err != nil {
			return nil, err
		}
		awayByIndex := make(map[int]*model.LineupPairing, len(awayPairings))
		for _, p := range awayPairings {
			awayByIndex[p.FormatLineIndex] = p
		}

		for _, hp := range homePairings {
			ap := awayByIndex[hp.FormatLineIndex]
			if ap == nil {
				continue
			}
			homeIM := &model.IndividualMatch{Structure: scoringStructureId, Status: model.MatchUnstarted}
			homeIMCreated, err := database.CreateOne(ctx, db, homeIM)
			if err != nil {
				return nil, err
			}
			awayIM := &model.IndividualMatch{Structure: scoringStructureId, Opponent: homeIMCreated.ID, Status: model.MatchUnstarted}
			awayIMCreated, err := database.CreateOne(ctx, db, awayIM)
			if err != nil {
				return nil, err
			}
			homeIMCreated.Opponent = awayIMCreated.ID
			if err := database.UpdateOne(ctx, db, homeIMCreated); err != nil {
				return nil, err
			}
			for _, c := range commissioners {
				if _, err := homeIMCreated.AssignEditor(ctx, db, c); err != nil {
					return nil, err
				}
				if _, err := awayIMCreated.AssignEditor(ctx, db, c); err != nil {
					return nil, err
				}
			}
			if _, err := database.CreateOne(ctx, db, &model.TeamMatchIndividualMatch{
				TeamMatchId:       tmCreated.ID,
				LineupPairingId:   hp.ID,
				IndividualMatchId: homeIMCreated.ID,
			}); err != nil {
				return nil, err
			}
			if _, err := database.CreateOne(ctx, db, &model.TeamMatchIndividualMatch{
				TeamMatchId:       tmCreated.ID,
				LineupPairingId:   ap.ID,
				IndividualMatchId: awayIMCreated.ID,
			}); err != nil {
				return nil, err
			}
		}

		tmDTO, err := buildTeamMatchDTO(ctx, db, tmCreated)
		if err != nil {
			return nil, err
		}
		detail.TeamMatches = append(detail.TeamMatches, tmDTO)
	}
	return detail, nil
}

// MatchQuery holds the query parameter for GetWeekMatches.
type MatchQuery struct {
	WeekId model.WeekId `json:"week_id"`
}

// StaticallyValid ensures a week is specified.
func (q *MatchQuery) StaticallyValid() error {
	if q.WeekId.RecordId() == database.InvalidRecordId {
		return errors.New("week_id must be set")
	}
	return nil
}

// GetWeekMatches returns the week's score sheet. It is viewable by everyone.
type GetWeekMatches struct{}

func (c GetWeekMatches) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, BaseRoute + "/week"
}

func (c GetWeekMatches) RequestBody() (*MatchQuery, bool) {
	return &MatchQuery{}, false
}

func (c GetWeekMatches) Handler(req api.Request[*MatchQuery]) (any, int, error) {
	query := req.HTTPRequest().URL.Query()
	weekStr := query.Get("week_id")
	if weekStr == "" {
		return nil, http.StatusBadRequest, errors.New("week_id must be set")
	}
	rid, err := database.RecordIdFromString(weekStr)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	week, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Week{}, rid)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	detail, err := weekDetailForWeek(req.Context, req.DatabaseProvider, week)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: detail}, http.StatusOK, nil
}

// ScoreBody is the request body for RecordScore.
type ScoreBody struct {
	IndividualMatchId string `json:"individual_match_id"`
	MainValue         int    `json:"main_value"`
	SecondaryValue    int    `json:"secondary_value"`
	WinOverride       bool   `json:"win_override"`
}

// StaticallyValid ensures an individual match is specified and scores are
// non-negative.
func (b *ScoreBody) StaticallyValid() error {
	if b.IndividualMatchId == "" {
		return errors.New("individual_match_id must be set")
	}
	if b.MainValue < 0 || b.SecondaryValue < 0 {
		return errors.New("scores cannot be negative")
	}
	return nil
}

// RecordScore updates the score on one side of an individual match. It does not
// complete the match; use CompleteMatch to determine the winner once both sides
// have been scored.
type RecordScore struct{}

func (c RecordScore) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute + "/score"
}

func (c RecordScore) RequestBody() (*ScoreBody, bool) {
	return &ScoreBody{}, true
}

func (c RecordScore) Handler(req api.Request[*ScoreBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}
	matchRid, err := database.RecordIdFromString(req.Body.IndividualMatchId)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	matchId := model.IndividualMatchId(matchRid)
	if !canEditMatch(req.Context, req.DatabaseProvider, req.Token.UserId, matchId) {
		return nil, http.StatusForbidden, errors.New("you are not an editor of this match")
	}
	im, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.IndividualMatch{}, matchId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	im.MainValue = req.Body.MainValue
	im.SecondaryValue = req.Body.SecondaryValue
	im.WinOverride = req.Body.WinOverride
	if err := database.UpdateOne(req.Context, req.DatabaseProvider, im); err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: im}, http.StatusOK, nil
}

// CompleteMatch marks an individual match complete, determining the winner
// against its opponent by the scoring structure once both sides have scores. A
// match with no opponent is recorded as won once it has a score.
type CompleteMatch struct{}

func (c CompleteMatch) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, api.AppendPathId(BaseRoute) + "/complete"
}

func (c CompleteMatch) RequestBody() (*EmptyBody, bool) {
	return &EmptyBody{}, false
}

func (c CompleteMatch) Handler(req api.Request[*EmptyBody]) (any, int, error) {
	if req.Token == nil {
		return nil, http.StatusUnauthorized, errors.New("token is required")
	}
	matchId := model.IndividualMatchId(req.PathId)
	if !canEditMatch(req.Context, req.DatabaseProvider, req.Token.UserId, matchId) {
		return nil, http.StatusForbidden, errors.New("you are not an editor of this match")
	}
	im, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.IndividualMatch{}, matchId.RecordId())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if err := determineWinnerAndMark(req.Context, req.DatabaseProvider, im); err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: im}, http.StatusOK, nil
}

// determineWinnerAndMark completes an individual match: when it has an
// opponent, both sides must be scored and the scoring structure decides the
// winner; the winner is marked Won and the loser Lost. A lone match (no
// opponent) is marked Won once it has a score.
func determineWinnerAndMark(ctx context.Context, db database.Provider, im *model.IndividualMatch) error {
	if im.Opponent == model.IndividualMatchId(database.InvalidRecordId) {
		if !hasScore(im) {
			return errors.New("match must be scored before completing")
		}
		im.Status = model.MatchWon
		return database.UpdateOne(ctx, db, im)
	}
	opp, err := database.GetExistingRecordById(ctx, db, &model.IndividualMatch{}, im.Opponent.RecordId())
	if err != nil {
		return err
	}
	if !hasScore(im) || !hasScore(opp) {
		return errors.New("both sides must be scored before completing the match")
	}
	if err := im.Initialize(ctx, db); err != nil {
		return err
	}
	if err := opp.Initialize(ctx, db); err != nil {
		return err
	}
	if im.Victorious(opp) {
		im.Status = model.MatchWon
		opp.Status = model.MatchLost
	} else if opp.Victorious(im) {
		opp.Status = model.MatchWon
		im.Status = model.MatchLost
	} else {
		return errors.New("match is not decided by the current scores")
	}
	if err := database.UpdateOne(ctx, db, opp); err != nil {
		return err
	}
	return database.UpdateOne(ctx, db, im)
}

// StandingsQuery holds the query parameter for GetStandings.
type StandingsQuery struct {
	SeasonId model.SeasonId `json:"season_id"`
}

// StaticallyValid ensures a season is specified.
func (q *StandingsQuery) StaticallyValid() error {
	if q.SeasonId.RecordId() == database.InvalidRecordId {
		return errors.New("season_id must be set")
	}
	return nil
}

// StandingsEntry is a team's cumulative win/loss/tie record across the season's
// completed team matches.
type StandingsEntry struct {
	TeamId string `json:"team_id"`
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Ties   int    `json:"ties"`
}

// GetStandings computes the season's weekly standings from completed team
// matches, sorted by wins (then ties). It is viewable by everyone.
type GetStandings struct{}

func (c GetStandings) Path() (api.HttpMethod, string) {
	return api.HttpMethodGet, BaseRoute + "/standings"
}

func (c GetStandings) RequestBody() (*StandingsQuery, bool) {
	return &StandingsQuery{}, false
}

func (c GetStandings) Handler(req api.Request[*StandingsQuery]) (any, int, error) {
	query := req.HTTPRequest().URL.Query()
	seasonStr := query.Get("season_id")
	if seasonStr == "" {
		return nil, http.StatusBadRequest, errors.New("season_id must be set")
	}
	rid, err := database.RecordIdFromString(seasonStr)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	season, err := database.GetExistingRecordById(req.Context, req.DatabaseProvider, &model.Season{}, rid)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	entries, err := computeStandings(req.Context, req.DatabaseProvider, season)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return gin.H{api.ResourceKey: entries}, http.StatusOK, nil
}

// computeStandings tallies each team's record across every completed team match
// in the season's weeks, sorted by wins then ties.
func computeStandings(ctx context.Context, db database.Provider, season *model.Season) ([]*StandingsEntry, error) {
	weeks, err := model.GetWeeksForDraft(ctx, db, season.DraftId)
	if err != nil {
		return nil, err
	}
	records := make(map[model.TeamId]*StandingsEntry)
	for _, week := range weeks {
		teamMatches, err := database.GetAllWhere[*model.TeamMatch](ctx, db, func(_ context.Context, tm *model.TeamMatch) bool {
			return tm.WeekId == week.ID
		})
		if err != nil {
			return nil, err
		}
		for _, tm := range teamMatches {
			dto, err := buildTeamMatchDTO(ctx, db, tm)
			if err != nil {
				return nil, err
			}
			if !dto.Complete {
				continue
			}
			if dto.Winner == tm.HomeTeam.RecordId().String() {
				bump(records, tm.HomeTeam, boolPtr(true))
				bump(records, tm.AwayTeam, boolPtr(false))
			} else if dto.Winner == tm.AwayTeam.RecordId().String() {
				bump(records, tm.AwayTeam, boolPtr(true))
				bump(records, tm.HomeTeam, boolPtr(false))
			} else {
				bump(records, tm.HomeTeam, nil)
				bump(records, tm.AwayTeam, nil)
			}
		}
	}
	out := make([]*StandingsEntry, 0, len(records))
	for _, e := range records {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Wins != out[j].Wins {
			return out[i].Wins > out[j].Wins
		}
		return out[i].Ties > out[j].Ties
	})
	return out, nil
}

// boolPtr returns a pointer to a bool literal (Go has no &true syntax).
func boolPtr(v bool) *bool { return &v }

// bump updates a team's win/loss/tie record. won is true for a win, false for a
// loss, and nil for a tie.
func bump(records map[model.TeamId]*StandingsEntry, teamId model.TeamId, won *bool) {
	e := records[teamId]
	if e == nil {
		e = &StandingsEntry{TeamId: teamId.RecordId().String()}
		records[teamId] = e
	}
	if won == nil {
		e.Ties++
	} else if *won {
		e.Wins++
	} else {
		e.Losses++
	}
}
