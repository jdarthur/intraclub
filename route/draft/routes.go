package draft

import (
	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up the Draft REST surface: standard CRUD for the draft
// and its join models (mirroring formats/ratings) plus the custom draft action
// endpoints.
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	drafts := api.NewCrudCommon(model.NewDraft, false, db)
	drafts.HandleRouteTypes(e, api.CrudWrapperFunctionAll...)

	availablePlayers := api.NewCrudCommon(func() *model.DraftAvailablePlayer { return &model.DraftAvailablePlayer{} }, false, db)
	availablePlayers.HandleRouteTypes(e, api.CrudWrapperFunctionAll...)

	captains := api.NewCrudCommon(func() *model.DraftCaptain { return &model.DraftCaptain{} }, false, db)
	captains.HandleRouteTypes(e, api.CrudWrapperFunctionAll...)

	picks := api.NewCrudCommon(func() *model.DraftPick { return &model.DraftPick{} }, false, db)
	picks.HandleRouteTypes(e, api.CrudWrapperFunctionAll...)

	ratingCutoffs := api.NewCrudCommon(func() *model.DraftRatingCutoff { return &model.DraftRatingCutoff{} }, false, db)
	ratingCutoffs.HandleRouteTypes(e, api.CrudWrapperFunctionAll...)

	preDraftGrades := api.NewCrudCommon(model.NewPreDraftGrade, false, db)
	preDraftGrades.HandleRouteTypes(e, api.CrudWrapperFunctionAll...)

	initFamily := api.RouteFamily[*InitializeBody]{DatabaseProvider: db}
	initFamily.Handle(e, InitializeDraft{})
	playersFamily := api.RouteFamily[*AssignDraftablePlayersBody]{DatabaseProvider: db}
	playersFamily.Handle(e, AssignDraftablePlayers{})
	cutoffFamily := api.RouteFamily[*AssignRatingCutoffBody]{DatabaseProvider: db}
	cutoffFamily.Handle(e, AssignRatingCutoff{})
	selectFamily := api.RouteFamily[*SelectBody]{DatabaseProvider: db}
	selectFamily.Handle(e, SelectByCaptain{})
	assignFamily := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	assignFamily.Handle(e, AssignDraftedPlayersToTeams{})
	seasonFamily := api.RouteFamily[*CreateSeasonBody]{DatabaseProvider: db}
	seasonFamily.Handle(e, CreateSeason{})
	patternFamily := api.RouteFamily[*SetDraftOrderPatternBody]{DatabaseProvider: db}
	patternFamily.Handle(e, SetDraftOrderPattern{})
}
