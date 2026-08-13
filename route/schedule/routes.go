package schedule

import (
	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires up the Schedule REST surface.
//
// Schedules and their weekly matchups are managed only by the Season
// commissioner; other participants can only view them. Creation and matchup
// assignment therefore go through the custom CreateSchedule /
// AssignWeeklyMatchup routes, which enforce the commissioner-only rule and
// write the sysadmin-owned join records (ScheduleMatchup,
// WeeklyMatchupTeamMatchup) on the commissioner's behalf. The underlying
// records are exposed read-only here.
func RegisterRoutes(e *gin.RouterGroup, db database.Provider) {
	create := api.RouteFamily[*CreateScheduleBody]{DatabaseProvider: db}
	create.Handle(e, CreateSchedule{})

	list := api.RouteFamily[*ScheduleQuery]{DatabaseProvider: db}
	list.Handle(e, ListSchedules{})

	detail := api.RouteFamily[*EmptyBody]{DatabaseProvider: db}
	detail.Handle(e, GetScheduleDetail{})

	assign := api.RouteFamily[*AssignWeeklyMatchupBody]{DatabaseProvider: db}
	assign.Handle(e, AssignWeeklyMatchup{})

	// Read-only generic surface for the underlying schedule records.
	weeklyMatchups := api.NewCrudCommon(model.NewWeeklyMatchup, false, db)
	weeklyMatchups.HandleRouteTypes(e, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)

	scheduleMatchups := api.NewCrudCommon(func() *model.ScheduleMatchup { return &model.ScheduleMatchup{} }, false, db)
	scheduleMatchups.HandleRouteTypes(e, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)

	weeklyMatchupTeamMatchups := api.NewCrudCommon(func() *model.WeeklyMatchupTeamMatchup { return &model.WeeklyMatchupTeamMatchup{} }, false, db)
	weeklyMatchupTeamMatchups.HandleRouteTypes(e, api.CrudWrapperFunctionGetOne, api.CrudWrapperFunctionGetMany)
}
