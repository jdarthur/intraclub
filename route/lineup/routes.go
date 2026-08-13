package lineup

import (
	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up the Lineup and LineupPairing REST surface.
//
// Generic CRUD is registered for both models: the models enforce that only a
// team's captain/co-captains can create/update/delete their lineup and its
// pairings (via EditableBy), and that pairings validate team membership and
// format ratings (via DynamicallyValid). Custom routes add the
// builder/confirm/official flow the season page uses:
//
//	GET  /lineup/detail?team_id=&week_id=   -> LineupDetail (lines, pairings)
//	POST /lineup/set                        -> build/replace a weekly lineup
//	POST /lineup/:id/confirm                -> captain confirms the lineup
//	POST /lineup/:id/official               -> commissioner marks it official
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	lineups := api.NewCrudCommon(model.NewLineup, false, db)
	lineups.HandleRouteTypes(e, api.CrudWrapperFunctionAll...)

	pairings := api.NewCrudCommon(model.NewLineupPairing, false, db)
	pairings.HandleRouteTypes(e, api.CrudWrapperFunctionAll...)

	detail := api.RouteFamily[*LineupQuery]{DatabaseProvider: db}
	detail.Handle(e, GetLineupDetail{})

	set := api.RouteFamily[*SetLineupBody]{DatabaseProvider: db}
	set.Handle(e, SetLineup{})

	confirm := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	confirm.Handle(e, ConfirmLineup{})

	official := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	official.Handle(e, MarkOfficial{})
}
