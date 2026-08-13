package week

import (
	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up the Week REST surface.
//
// Weeks are created and managed only by the Season commissioner. Generic
// create is deliberately NOT registered: creation goes through the custom
// CreateWeek route, which enforces the commissioner-only rule. The generic
// get/update/delete routes enforce the same rule via Week.EditableBy (which
// returns the Season's commissioners).
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	weeks := api.NewCrudCommon(model.NewWeek, false, db)
	weeks.HandleRouteTypes(e,
		api.CrudWrapperFunctionGetOne,
		api.CrudWrapperFunctionUpdate,
		api.CrudWrapperFunctionDelete,
	)

	queryFamily := api.RouteFamily[*WeekQuery]{DatabaseProvider: db}
	queryFamily.Handle(e, ListWeeks{})

	createFamily := api.RouteFamily[*CreateWeekBody]{DatabaseProvider: db}
	createFamily.Handle(e, CreateWeek{})
}
