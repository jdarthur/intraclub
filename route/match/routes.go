package match

import (
	"intraclub/api"
	"intraclub/database"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up the Match REST surface.
//
// Match scoring is a constrained flow layered on top of the schedule and
// lineup builders: the season commissioner generates a week's TeamMatches from
// the scheduled weekly matchups and both teams' official lineups, editors
// record individual-match scores, and the commissioner closes the week once
// every team match is complete. The underlying records (IndividualMatch,
// TeamMatch, TeamMatchIndividualMatch, MatchEditor) are also exposed via
// generic CRUD in main.go.
//
//	POST /match/generate     body: { week_id, scoring_structure_id }
//	GET  /match/week?week_id=              -> WeekMatchDetail (score sheet)
//	POST /match/score        body: { individual_match_id, main_value, secondary_value, win_override }
//	POST /match/:id/complete -> mark an individual match complete (determines winner)
//	GET  /match/standings?season_id=       -> Standings
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	generate := api.RouteFamily[*GenerateBody]{DatabaseProvider: db}
	generate.Handle(e, GenerateMatches{})

	week := api.RouteFamily[*MatchQuery]{DatabaseProvider: db}
	week.Handle(e, GetWeekMatches{})

	score := api.RouteFamily[*ScoreBody]{DatabaseProvider: db}
	score.Handle(e, RecordScore{})

	complete := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	complete.Handle(e, CompleteMatch{})

	standings := api.RouteFamily[*StandingsQuery]{DatabaseProvider: db}
	standings.Handle(e, GetStandings{})
}
