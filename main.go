package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"intraclub/route"
)

func main() {
	common.SysAdminCheck = model.IsUserSystemAdministrator
	common.UserType = &model.User{}

	// set up the default database provider
	common.GlobalDatabaseProvider = common.NewUnitTestDBProvider()

	// parse command-line flags
	parseFlags()

	// seed users for development mode
	seedDevUsers()

	// generate or load JWT key pair
	err := common.GenerateJwtKeyPairIfNotExists()
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	api := r.Group("/api")

	// noAuth for self-register
	createUser := common.RouteFamily[*model.User]{}
	createUser.Handle(api, route.SelfRegister{})

	whoAmI := common.RouteFamily[*model.User]{UseAuth: true}
	whoAmI.Handle(api, route.WhoAmI{})

	startTokenMgr := &model.StartLoginTokenManager{}
	api.POST("/one_time_password", startTokenMgr.OneTimePassword)
	api.POST("/token", startTokenMgr.CreateJwtFromOneTimePassword)

	// no auth for get user by ID / get all users functions

	getUsers := common.NewCrudCommon(model.NewUser, false)
	getUsers.HandleRouteTypes(api, common.CrudWrapperFunctionGetOne, common.CrudWrapperFunctionGetMany)

	// use auth for user deletion / update endpoints
	updateOrDeleteUsers := common.NewCrudCommon(model.NewUser, true)
	updateOrDeleteUsers.HandleRouteTypes(api, common.CrudWrapperFunctionDelete, common.CrudWrapperFunctionUpdate)

	facilities := common.NewCrudCommon(model.NewFacility, true)
	facilities.HandleRouteTypes(api, common.CrudWrapperFunctionAll...)

	err = r.Run("127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
}

func parseFlags() {
	useDevTokenMode := flag.Bool("dev-token", false, "Use development token mode")
	flag.Parse()

	if useDevTokenMode != nil && *useDevTokenMode == true {
		model.UseDevTokenMode = true
		fmt.Println("Using development token mode")
	}
}

func seedDevUsers() {
	user1 := &model.User{
		FirstName: "JD",
		LastName:  "Arthur",
		Email:     "jdarthur@gatech.edu",
	}

	_, err := common.CreateOne(common.GlobalDatabaseProvider, user1)
	if err != nil {
		panic(err)
	}
}
