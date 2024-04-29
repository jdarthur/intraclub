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

	leagueCtl := common.CrudController{Controller: controllers.LeagueController{}, Database: common.GlobalDbProvider}
	apiAuthAndAccess.Handle("GET", "/leagues", leagueCtl.GetAll)

	// get league by ID is a no-auth route so you can view basic league info without being logged in
	noAuth.Handle("GET", "/league/:id", leagueCtl.GetOne)

	apiAuthAndAccess.Handle("GET", "/whoami", controllers.WhoAmI)

	noAuth.Handle("GET", "/teams_for_user/:id", controllers.GetTeamsForUser)
	noAuth.Handle("GET", "/leagues_for_user/:id", controllers.GetLeaguesForUser)

	facilityCtl := common.CrudController{Controller: controllers.FacilityController{}, Database: common.GlobalDbProvider}

	apiAuthAndAccess.Handle("GET", "/facilities", facilityCtl.GetAll)
	apiAuthAndAccess.Handle("POST", "/facilities", facilityCtl.Create)

	// get facility by ID without authentication
	noAuth.Handle("GET", "/facilities/:id", facilityCtl.GetOne)

	facilityOwnedByUser := middleware.OwnedByUserWrapper{Record: &model.Facility{}}
	apiAuthAndAccess.Handle("DELETE", "/facilities/:id", facilityOwnedByUser.OwnedByUser, facilityCtl.Delete)
	apiAuthAndAccess.Handle("PUT", "/facilities/:id", facilityOwnedByUser.OwnedByUser, facilityCtl.Update)

	apiAuthAndAccess.Handle("GET", "/leagues_commissioned_by_user/:id", controllers.GetCommissionedLeaguesForUser)

	err = router.Run(":8080")
	if err != nil {
		panic(err)
	}

}
