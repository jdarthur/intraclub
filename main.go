package main

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/controllers"
	"intraclub/middleware"
	"intraclub/model"
	"sync"
)

func main() {
	common.GlobalDbProvider = model.NewMongoDbProvider("mongodb://localhost:27018", "", "")
	err := common.GlobalDbProvider.Connect()
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	noAuth := router.Group("/api")
	apiAuthAndAccess := router.Group("/api", middleware.WithToken)

	otpManager := controllers.OneTimePasswordManager{Map: &sync.Map{}}
	noAuth.POST("/token", otpManager.GetToken)
	noAuth.POST("/one_time_password", otpManager.Create)

	registerManager := controllers.NewUserController{}
	noAuth.POST("/register", registerManager.Register)

	userCtl := common.CrudController{Controller: controllers.UserController{}, Database: common.GlobalDbProvider}
	apiAuthAndAccess.Handle("GET", "/users", userCtl.GetAll)
	noAuth.Handle("GET", "/users/:id", userCtl.GetOne)

	// you can only delete your own user
	ownedByUser := middleware.OwnedByUserWrapper{Record: &model.User{}}
	apiAuthAndAccess.Handle("DELETE", "/users/:id", ownedByUser.OwnedByUser, userCtl.Delete)

	// special route that we can use to retrieve the active user from a token
	apiAuthAndAccess.Handle("GET", "/whoami", controllers.WhoAmI)

	// league CRUD controller
	leagueCtl := common.CrudController{Controller: controllers.LeagueController{}, Database: common.GlobalDbProvider}
	apiAuthAndAccess.Handle("GET", "/leagues", leagueCtl.GetAll)
	apiAuthAndAccess.Handle("POST", "/leagues", leagueCtl.Create)

	// get league by ID is a no-auth route so you can view basic league info without being logged in
	noAuth.Handle("GET", "/league/:id", leagueCtl.GetOne)
	apiAuthAndAccess.Handle("GET", "/league/:id/weeks", controllers.GetWeeksForLeague)

	leagueOwnedByUser := middleware.OwnedByUserWrapper{Record: &model.League{}}
	apiAuthAndAccess.Handle("DELETE", "/leagues/:id", leagueOwnedByUser.OwnedByUser, leagueCtl.Delete)
	apiAuthAndAccess.Handle("PUT", "/leagues/:id", leagueOwnedByUser.OwnedByUser, leagueCtl.Update)

	// Get relevant records by user ID
	noAuth.Handle("GET", "/teams_for_user/:id", controllers.GetTeamsForUser)
	noAuth.Handle("GET", "/leagues_for_user/:id", controllers.GetLeaguesForUser)
	apiAuthAndAccess.Handle("GET", "/leagues_commissioned_by_user/:id", controllers.GetCommissionedLeaguesForUser)

	// facility CRUD controller
	facilityCtl := common.CrudController{Controller: controllers.FacilityController{}, Database: common.GlobalDbProvider}
	apiAuthAndAccess.Handle("GET", "/facilities", facilityCtl.GetAll)
	apiAuthAndAccess.Handle("POST", "/facilities", facilityCtl.Create)

	// get facility by ID without authentication
	noAuth.Handle("GET", "/facilities/:id", facilityCtl.GetOne)

	// Delete / Update facility endpoints are protected by OwnedByUser middleware
	facilityOwnedByUser := middleware.OwnedByUserWrapper{Record: &model.Facility{}}
	apiAuthAndAccess.Handle("DELETE", "/facilities/:id", facilityOwnedByUser.OwnedByUser, facilityCtl.Delete)
	apiAuthAndAccess.Handle("PUT", "/facilities/:id", facilityOwnedByUser.OwnedByUser, facilityCtl.Update)

	// week CRUD controller
	weekCtl := common.CrudController{Controller: controllers.WeekController{}, Database: common.GlobalDbProvider}
	apiAuthAndAccess.Handle("POST", "/weeks", weekCtl.Create)
	apiAuthAndAccess.Handle("GET", "/weeks/:id", weekCtl.GetOne)
	apiAuthAndAccess.Handle("POST", "/weeks_search", controllers.GetWeeksByIds)

	weekOwnedByUser := middleware.OwnedByUserWrapper{Record: &model.Week{}}
	apiAuthAndAccess.Handle("DELETE", "/weeks/:id", weekOwnedByUser.OwnedByUser, weekCtl.Delete)

	apiAuthAndAccess.Handle("POST", "import_users_from_csv", controllers.ParseUserCsv)

	noAuth.Handle("GET", "/match_scores", controllers.GetMatchScoreboard)
	noAuth.Handle("PUT", "/match_scores", controllers.UpdateMatchScore)
	noAuth.Handle("PUT", "/match_player_names", controllers.UpdateMatchNames)
	noAuth.Handle("PUT", "/match_team_info", controllers.UpdateMatchTeam)

	// initialize the websocket server + serve a websocket connection func on the /api/ws path.
	// this is used for notifications on the client side as opposed to a polling loop which
	// can be kind of wasteful if we have a lot of clients connected or a client connected for
	// a long amount of time without any updates actually occurring to the scoreboard
	controllers.WebsocketInit()
	noAuth.Handle("GET", "/ws", controllers.WebSocketServer)

	err = router.Run(":8080")
	if err != nil {
		panic(err)
	}

}
