package team

import (
	"intraclub/api"
	"intraclub/database"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up the constrained Team REST surface.
//
// Team and TeamAssignment records are exposed read-only (GET one / GET many,
// registered in main.go) — there is deliberately no generic create/update/delete
// on the raw records, since rosters are fixed at draft finalize time. This
// function adds the single mutation endpoint, co-captain role assignment.
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	rosterFamily := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	rosterFamily.Handle(e, ListTeams{}, GetTeam{})

	promoteFamily := api.RouteFamily[*PromoteCoCaptainBody]{DatabaseProvider: db}
	promoteFamily.Handle(e, PromoteCoCaptain{})
}
